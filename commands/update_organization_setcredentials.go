package commands

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/giantswarm/gsclientgen/models"
	"github.com/giantswarm/gsctl/client/clienterror"
	"github.com/giantswarm/gsctl/config"
	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"
)

const (
	updateOrgSetCredentialsActivityName = "set-org-credentials"
)

var (

	// UpdateOrgSetCredentialsCommand performs the "update organizatio set-credentials" function
	UpdateOrgSetCredentialsCommand = &cobra.Command{
		Use:     "set-credentials",
		Aliases: []string{"sc"},
		Short:   "Set credentials of an organization",
		Long: `Set the credentials used to create and operate the clusters of an organization.

Setting credentials of an organization will result in all future cluster
being run in the account/subscription referenced by the credentials. Once
credentials are set for an organization, this cannot be undone.

For details on how to prepare the account/subscription, consult the documentation at

  - https://docs.giantswarm.io/guides/prepare-aws-account-for-tenant-clusters/ (AWS)
  - https://docs.giantswarm.io/guides/prepare-azure-subscription-for-tenant-clusters/ (Azure)

`,
		Example: `
  gsctl update organization set-credentials -o acme \
    --aws-operator-role arn:aws:iam::<AWS-ACCOUNT-ID>:role/GiantSwarmAWSOperator \
    --aws-admin-role arn:aws:iam::<AWS-ACCOUNT-ID>:role/GiantSwarmAdmin

  gsctl update organization set-credentials -o acme \
    --azure-subscription-id <AZURE-SUBSCRIPTION-ID> \
    --azure-tenant-id <AZURE-TENANT-ID> \
    --azure-client-id <AZURE-CLIENT-ID> \
    --azure-secret-key <AZURE-SECRET-KEY>
`,

		// PreRun checks a few general things, like authentication and flags
		// plausibility.
		PreRun: updateOrgSetCredentialsPreRunOutput,

		// Run calls the business function and prints results and errors.
		Run: updateOrgSetCredentialsRunOutput,
	}

	// AWS operator role ARN as passed by the user via flags
	cmdAWSOperatorRoleARN string

	// AWS admin role ARN as passed by the user via flags
	cmdAWSAdminRoleARN string

	// Azure-related flags
	cmdAzureSubscriptionID string
	cmdAzureTenantID       string
	cmdAzureClientID       string
	cmdAzureSecretKey      string

	// Here we briefly store the info which provider we are dealing with
	provider string
)

type updateOrgSetCredentialsArguments struct {
	apiEndpoint         string
	authToken           string
	scheme              string
	verbose             bool
	organizationID      string
	awsAdminRole        string
	awsOperatorRole     string
	azureSubscriptionID string
	azureTenantID       string
	azureClientID       string
	azureSecretKey      string
}

type updateOrgSetCredentialsResult struct {
	credentialID string
}

func init() {
	UpdateOrgSetCredentialsCommand.Flags().StringVarP(&cmdOrganization, "organization", "o", "", "ID of the organization to set credentials for")
	UpdateOrgSetCredentialsCommand.Flags().StringVarP(&cmdAWSOperatorRoleARN, "aws-operator-role", "", "", "AWS ARN of the role to use for operating clusters")
	UpdateOrgSetCredentialsCommand.Flags().StringVarP(&cmdAWSAdminRoleARN, "aws-admin-role", "", "", "AWS ARN of the role to be used by Giant Swarm staff")
	UpdateOrgSetCredentialsCommand.Flags().StringVarP(&cmdAzureSubscriptionID, "azure-subscription-id", "", "", "ID of the Azure subscription to run clusters in")
	UpdateOrgSetCredentialsCommand.Flags().StringVarP(&cmdAzureTenantID, "azure-tenant-id", "", "", "ID of the Azure tenant to run clusters in")
	UpdateOrgSetCredentialsCommand.Flags().StringVarP(&cmdAzureClientID, "azure-client-id", "", "", "ID of the Azure service principal to use for operating clusters")
	UpdateOrgSetCredentialsCommand.Flags().StringVarP(&cmdAzureSecretKey, "azure-secret-key", "", "", "Secret key for the Azure service principal to use for operating clusters")

	UpdateOrganizationCommand.AddCommand(UpdateOrgSetCredentialsCommand)
}

func defaultUpdateOrgSetCredentialsArguments() updateOrgSetCredentialsArguments {
	endpoint := config.Config.ChooseEndpoint(cmdAPIEndpoint)
	token := config.Config.ChooseToken(endpoint, cmdToken)
	scheme := config.Config.ChooseScheme(endpoint, cmdToken)

	return updateOrgSetCredentialsArguments{
		apiEndpoint:         endpoint,
		authToken:           token,
		scheme:              scheme,
		organizationID:      cmdOrganization,
		verbose:             cmdVerbose,
		awsAdminRole:        cmdAWSAdminRoleARN,
		awsOperatorRole:     cmdAWSOperatorRoleARN,
		azureClientID:       cmdAzureClientID,
		azureSecretKey:      cmdAzureSecretKey,
		azureSubscriptionID: cmdAzureSubscriptionID,
		azureTenantID:       cmdAzureTenantID,
	}
}

func updateOrgSetCredentialsPreRunOutput(cmd *cobra.Command, cmdLineArgs []string) {
	args := defaultUpdateOrgSetCredentialsArguments()
	err := verifyUpdateOrgSetCredentialsPreconditions(args)

	if err == nil {
		return
	}

	handleCommonErrors(err)

	// From here on we handle errors that can only occur in this command
	headline := ""
	subtext := ""

	switch {
	case err.Error() == "":
		return
	case IsOrganizationNotSpecifiedError(err):
		headline = "No organization given"
		subtext = "Please specify the organization to set credentials for using the -o|--organization flag."
	case IsProviderNotSupportedError(err):
		headline = "Unsupported provider"
		subtext = "Setting credentials is only supported on AWS and Azure installations."
	case IsRequiredFlagMissingError(err):
		headline = "Missing flag: " + err.Error()
		subtext = "Please use --help to see details regarding the command's usage."
	case IsConflictingFlagsError(err):
		headline = "Conflicting flags"
		subtext = "Please use only AWS or Azure related flags with this installation. See --help for details."
	case IsOrganizationNotFoundError(err):
		headline = fmt.Sprintf("Organization '%s' not found", args.organizationID)
		subtext = "The specified organization does not exist, or you are not a member. Please check the exact upper-/lower case spelling."
		subtext += "\nUse 'gsctl list organizations' to list all organizations."
	default:
		headline = err.Error()
	}

	// print output
	fmt.Println(color.RedString(headline))
	if subtext != "" {
		fmt.Println(subtext)
	}
	os.Exit(1)
}

func verifyUpdateOrgSetCredentialsPreconditions(args updateOrgSetCredentialsArguments) error {
	if args.organizationID == "" {
		return microerror.Mask(organizationNotSpecifiedError)
	}
	if config.Config.Token == "" && args.authToken == "" {
		return microerror.Mask(notLoggedInError)
	}

	// get installation's provider (supported: aws, azure)
	auxParams := ClientV2.DefaultAuxiliaryParams()
	auxParams.ActivityName = updateOrgSetCredentialsActivityName
	response, err := ClientV2.GetInfo(auxParams)
	if err != nil {
		if clientErr, ok := err.(*clienterror.APIError); ok {
			if clientErr.HTTPStatusCode == http.StatusUnauthorized {
				return microerror.Mask(notAuthorizedError)
			} else if clientErr.HTTPStatusCode == http.StatusForbidden {
				return microerror.Mask(accessForbiddenError)
			}
		}
		return microerror.Mask(err)
	}

	provider = response.Payload.General.Provider

	if provider != "aws" && provider != "azure" {
		return microerror.Mask(providerNotSupportedError)
	}

	// check flags based on provider
	{
		if provider == "aws" {
			if args.awsAdminRole == "" {
				return microerror.Maskf(requiredFlagMissingError, "--aws-admin-role")
			}
			if args.awsOperatorRole == "" {
				return microerror.Maskf(requiredFlagMissingError, "--aws-operator-role")
			}

			// conflicts
			if args.azureClientID != "" || args.azureSecretKey != "" || args.azureSubscriptionID != "" || args.azureTenantID != "" {
				return microerror.Maskf(conflictingFlagsError, "Azure-related flags not allowed here")
			}
		}
		if provider == "azure" {
			if args.azureClientID == "" {
				return microerror.Maskf(requiredFlagMissingError, "--azure-client-id")
			}
			if args.azureSecretKey == "" {
				return microerror.Maskf(requiredFlagMissingError, "--azure-secret-key")
			}
			if args.azureSubscriptionID == "" {
				return microerror.Maskf(requiredFlagMissingError, "--azure-subscription-id")
			}
			if args.azureTenantID == "" {
				return microerror.Maskf(requiredFlagMissingError, "--azure-tenant-id")
			}

			// conflicts
			if args.awsAdminRole != "" || args.awsOperatorRole != "" {
				return microerror.Maskf(conflictingFlagsError, "AWS-related flags not allowed here")
			}
		}
	}

	// check organization membership and existence
	orgsResponse, err := ClientV2.GetOrganizations(auxParams)
	{
		if err != nil {
			if clientErr, ok := err.(*clienterror.APIError); ok {
				if clientErr.HTTPStatusCode == http.StatusUnauthorized {
					return microerror.Mask(notAuthorizedError)
				} else if clientErr.HTTPStatusCode == http.StatusForbidden {
					return microerror.Mask(accessForbiddenError)
				}
			}
			return microerror.Mask(err)
		}

		foundOrg := false
		for _, org := range orgsResponse.Payload {
			if org.ID == args.organizationID {
				foundOrg = true
			}
		}
		if !foundOrg {
			return microerror.Mask(organizationNotFoundError)
		}
	}

	return nil
}

// updateOrgSetCredentialsRunOutput calls the busniness function and produces
// meanigful terminal output.
func updateOrgSetCredentialsRunOutput(cmd *cobra.Command, cmdLineArgs []string) {
	args := defaultUpdateOrgSetCredentialsArguments()
	result, err := updateOrgSetCredentials(args)

	if err != nil {
		handleCommonErrors(err)

		// From here on we handle errors that can only occur in this command
		headline := ""
		subtext := ""

		switch {
		case err.Error() == "":
			return
		case IsOrganizationNotSpecifiedError(err):
			headline = "No organization given"
			subtext = "Please specify the organization to set credentials for using the -o|--organization flag."
		default:
			headline = err.Error()
		}

		// print output
		fmt.Println(color.RedString(headline))
		if subtext != "" {
			fmt.Println(subtext)
		}
		os.Exit(1)
	}

	// success
	fmt.Println(color.GreenString("Credentials set successfully"))
	fmt.Printf("The credentials are stored with the unique ID '%s'.\n", result.credentialID)
}

// updateOrgSetCredentials performs the API call and provides a result.
func updateOrgSetCredentials(args updateOrgSetCredentialsArguments) (*updateOrgSetCredentialsResult, error) {
	// build request body based on provider
	requestBody := &models.V4AddCredentialsRequest{Provider: &provider}
	if provider == "aws" {
		requestBody.Aws = &models.V4AddCredentialsRequestAws{
			Roles: &models.V4AddCredentialsRequestAwsRoles{
				Admin:       &args.awsAdminRole,
				Awsoperator: &args.awsOperatorRole,
			},
		}
	} else if provider == "azure" {
		requestBody.Azure = &models.V4AddCredentialsRequestAzure{
			Credential: &models.V4AddCredentialsRequestAzureCredential{
				SubscriptionID: &args.azureSubscriptionID,
				TenantID:       &args.azureTenantID,
				ClientID:       &args.azureClientID,
				SecretKey:      &args.azureSecretKey,
			},
		}
	}

	auxParams := ClientV2.DefaultAuxiliaryParams()
	auxParams.ActivityName = updateOrgSetCredentialsActivityName
	response, err := ClientV2.SetCredentials(args.organizationID, requestBody, auxParams)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// Location header returned is in the format
	// /v4/organizations/myorg/credentials/{credential_id}/
	segments := strings.Split(response.Location, "/")
	result := &updateOrgSetCredentialsResult{
		credentialID: segments[len(segments)-2],
	}

	return result, nil
}
