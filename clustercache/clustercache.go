package clustercache

import (
	"fmt"
	"path"
	"time"

	"github.com/fatih/color"
	"github.com/giantswarm/columnize"
	"github.com/giantswarm/gscliauth/config"
	"github.com/giantswarm/gsclientgen/models"
	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/client/clienterror"
	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/confirm"
	"github.com/giantswarm/microerror"
	"github.com/spf13/afero"
	yaml "gopkg.in/yaml.v2"
)

const (
	listClustersActivityName = "list-clusters"
	clusterCacheFileName     = "clustercache.yaml"

	cacheDuration = time.Hour * 24 * 7 // 7 days.
	timeLayout    = time.RFC3339
)

// EndpointCache stores the IDs stored in an
// endpoint-specific cache, and also its expiry date.
type EndpointCache struct {
	Expiry string   `yaml:"expiry"`
	IDs    []string `yaml:"ids"`
}

// Endpoints stores a map with the keys being API endpoints,
// and the values being caches.
type Endpoints map[string]EndpointCache

// Cache is the file structure of the cluster cache file.
type Cache struct {
	Endpoints Endpoints `yaml:"endpoints"`
}

// GetID gets the cluster ID for a provided name/ID
// by checking in both the user cache and on the API.
func GetID(endpoint string, clusterNameOrID string, clientWrapper *client.Wrapper) (string, error) {
	// Check if the cluster ID is already in the cache,
	// and skip the API request if it is.
	isInCache := IsInCache(endpoint, clusterNameOrID)
	if isInCache {
		return clusterNameOrID, nil
	}

	auxParams := clientWrapper.DefaultAuxiliaryParams()
	auxParams.ActivityName = listClustersActivityName

	// Get a list of all clusters.
	response, err := clientWrapper.GetClusters(auxParams)
	if err != nil {
		switch {
		case clienterror.IsUnauthorizedError(err):
			return "", microerror.Mask(errors.NotAuthorizedError)

		case clienterror.IsAccessForbiddenError(err):
			return "", microerror.Mask(errors.AccessForbiddenError)

		default:
			return "", microerror.Mask(err)
		}
	}

	var (
		// IDs that correspond to the same cluster name.
		matchingIDs []string
		// All IDs returned from the response.
		allClusterIDs = make([]string, 0, len(response.Payload))
	)
	for _, cluster := range response.Payload {
		allClusterIDs = append(allClusterIDs, cluster.ID)
		// Check if this is the cluster we're looking for.
		if matchesValidation(clusterNameOrID, cluster) {
			matchingIDs = append(matchingIDs, cluster.ID)
		}
	}

	CacheIDs(endpoint, allClusterIDs)

	if matchingIDs == nil {
		// There are no IDs that correspond to that cluster name
		return "", microerror.Mask(errors.ClusterNotFoundError)
	} else if len(matchingIDs) > 1 {
		// There are multiple IDs that correspond to that cluster name.
		// Help the user decide which one to pick.
		id := handleNameCollision(clusterNameOrID, response.Payload)

		return id, nil
	}

	return matchingIDs[0], nil
}

// New creates a new Cache object.
func New() *Cache {
	c := &Cache{}
	c.Endpoints = Endpoints{}

	return c
}

// IsInCache checks if a cluster ID is present in the
// persistent cluster cache.
func IsInCache(endpoint string, ID string) bool {
	var (
		err                error
		endpointExpiration time.Time

		now = time.Now()
	)
	// Read the existing cache file.
	existing, err := read(config.FileSystem)
	if err != nil {
		return false
	}

	c := existing.Endpoints[endpoint]
	for _, cID := range c.IDs {
		// Check if the cache is expired.
		endpointExpiration, err = time.Parse(timeLayout, c.Expiry)
		if err != nil || now.After(endpointExpiration) {
			return false
		}
		// This is the one we're looking for.
		if cID == ID {
			return true
		}
	}

	return false
}

// CacheIDs adds cluster IDs to a persistent cache,
// which can be used for decreasing timeout in getting
// cluster IDs, for commands that take both cluster names and IDs.
func CacheIDs(endpoint string, c []string) {
	// Let's not store an empty list.
	if len(c) == 0 {
		return
	}

	fs := config.FileSystem

	var cache *Cache
	{
		// Create a new Cache object if there is no file there yet.
		cache, _ = read(fs)
		if cache == nil {
			cache = New()
		}
	}

	// Add the cache to a certain endpoint.
	cache.Endpoints[endpoint] = EndpointCache{
		Expiry: time.Now().Add(cacheDuration).Format(timeLayout),
		IDs:    c,
	}

	// Write the cache file.
	err := write(fs, cache)
	if err != nil {
		return
	}
}

func matchesValidation(nameOrID string, cluster *models.V4ClusterListItem) bool {
	return cluster.DeleteDate == nil && (cluster.ID == nameOrID || cluster.Name == nameOrID)
}

func handleNameCollision(nameOrID string, clusters []*models.V4ClusterListItem) string {
	var (
		clusterIDs []string

		table = []string{color.CyanString("ID | ORGANIZATION | NAME")}
	)

	for _, cluster := range clusters {
		if matchesValidation(nameOrID, cluster) {
			clusterIDs = append(clusterIDs, cluster.ID)
			table = append(table, fmt.Sprintf("%5s | %5s | %5s\n", cluster.ID, cluster.Owner, cluster.Name))
		}
	}

	printNameCollisionTable(table)

	confirmed, id := confirm.AskStrictOneOf(
		fmt.Sprintf(
			"Found more than one cluster called '%s', please type the ID of the cluster that you would like to delete",
			nameOrID,
		),
		clusterIDs,
	)
	if !confirmed {
		return ""
	}
	return id
}

func printNameCollisionTable(table []string) {
	// Print output
	fmt.Println("Multiple clusters found")
	fmt.Printf("\n")
	fmt.Println(columnize.SimpleFormat(table))
	fmt.Printf("\n")
}

func read(fs afero.Fs) (*Cache, error) {
	cache := New()

	filePath := path.Join(config.ConfigDirPath, clusterCacheFileName)
	yamlBytes, err := afero.ReadFile(fs, filePath)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(yamlBytes, cache)
	if err != nil {
		return nil, err
	}

	return cache, nil
}

func write(fs afero.Fs, c *Cache) error {
	filePath := path.Join(config.ConfigDirPath, clusterCacheFileName)
	output, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	err = afero.WriteFile(fs, filePath, output, config.ConfigFilePermission)

	return err
}
