package commands

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/fatih/color"
	apischema "github.com/giantswarm/api-schema"
	"github.com/giantswarm/gsclientgen"
	"github.com/howeyc/gopass"
	keychain "github.com/lunixbochs/go-keychain"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/config"
)

const (
	loginActivityName string = "login"

	// to store how the user has provided the password
	passwordEntryMethodUnknown  int32 = 0
	passwordEntryMethodFlag     int32 = 1
	passwordEntryMethodPrompt   int32 = 2
	passwordEntryMethodKeychain int32 = 3
)

var (
	// password given via command line flag
	password string

	// how the password was provided
	passwordEntryMethod int32

	// LoginCommand is the "login" CLI command
	LoginCommand = &cobra.Command{
		Use:     "login <email>",
		Short:   "Sign in as a user",
		Long:    `Sign in with email address and password. Password has to be entered interactively or given as -p flag.`,
		PreRunE: checkLogin,
		Run:     login,
	}
)

func init() {
	LoginCommand.Flags().StringVarP(&password, "password", "p", "", "Password. If not given, will be prompted interactively.")
	RootCommand.AddCommand(LoginCommand)
}

// checks if all arguments for the login command are given
func checkLogin(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New(color.RedString("The email argument is required"))
	}

	// using auth token flag?
	if cmdToken != "" {
		return errors.New(color.RedString("The 'login' command cannot be used with the '--auth-token' flag"))
	}

	// already logged in?
	if config.Config.Token != "" {
		return errors.New(color.RedString("You are already logged in"))
	}

	return nil
}

// login creates a new session token
func login(cmd *cobra.Command, args []string) {
	email := args[0]
	keychainServiceName := "gsctl password for " + email

	if password != "" {
		// given via command line
		passwordEntryMethod = passwordEntryMethodFlag
	}

	// find keychain entry
	if runtime.GOOS == "darwin" {
		foundPassword, findErr := keychain.Find(keychainServiceName, email)
		if findErr == nil {
			if cmdVerbose {
				fmt.Println("Using password from Keychain")
			}
			password = foundPassword
			passwordEntryMethod = passwordEntryMethodKeychain
		}
	}

	if password == "" {
		// interactive password prompt
		fmt.Printf("Password: ")
		pass, err := gopass.GetPasswd()
		if err != nil {
			log.Fatal(err)
		}
		password = string(pass)
		passwordEntryMethod = passwordEntryMethodPrompt
	}

	encodedPassword := base64.StdEncoding.EncodeToString([]byte(password))

	client := gsclientgen.NewDefaultApiWithBasePath(cmdAPIEndpoint)
	requestBody := gsclientgen.LoginBodyModel{Password: string(encodedPassword)}
	loginResponse, apiResponse, err := client.UserLogin(email, requestBody, requestIDHeader, loginActivityName, cmdLine)
	if err != nil {
		log.Fatal(err)
	}
	if loginResponse.StatusCode == apischema.STATUS_CODE_DATA {
		// successful login
		config.Config.Token = loginResponse.Data.Id
		config.Config.Email = email
		fmt.Println(color.GreenString("Successfully logged in"))
		// store password in keychain
		if runtime.GOOS == "darwin" && passwordEntryMethod == passwordEntryMethodPrompt {
			c := askForConfirmation("Do you want to store this password to the Mac OS Keychain to speed up future logins?")
			if c {
				keychain.Add(keychainServiceName, email, password)
				if cmdVerbose {
					fmt.Println("Password stored to Keychain")
				}
			}
		}
	} else if loginResponse.StatusCode == apischema.STATUS_CODE_RESOURCE_INVALID_CREDENTIALS {
		// bad credentials
		fmt.Println(color.RedString("Incorrect password submitted. Please try again."))
		os.Exit(1)
	} else if loginResponse.StatusCode == apischema.STATUS_CODE_WRONG_INPUT {
		// empty password
		fmt.Println(color.RedString("The password must not be empty. Please try again."))
		os.Exit(1)
	} else {
		fmt.Printf("Unhandled response code: %v", loginResponse.StatusCode)
		fmt.Printf("Status text: %v", loginResponse.StatusText)
		fmt.Printf("apiResponse: %s\n", apiResponse)
		os.Exit(1)
	}

	config.WriteToFile()
}
