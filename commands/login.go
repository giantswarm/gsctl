package commands

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/fatih/color"
	apischema "github.com/giantswarm/api-schema"
	"github.com/giantswarm/gsclientgen"
	"github.com/howeyc/gopass"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/config"
)

const (
	loginActivityName string = "login"
)

var (
	// password given via command line flag
	password string

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
	var email = args[0]

	// interactive password prompt
	if password == "" {
		fmt.Printf("Password: ")
		pass, err := gopass.GetPasswd()
		if err != nil {
			log.Fatal(err)
		}
		password = string(pass)
	}

	encodedPassword := base64.StdEncoding.EncodeToString([]byte(password))

	client := gsclientgen.NewDefaultApi()
	requestBody := gsclientgen.LoginBodyModel{Password: string(encodedPassword)}
	loginResponse, apiResponse, err := client.UserLogin(email, requestBody, requestIDHeader, loginActivityName, cmdLine)
	if err != nil {
		log.Fatal(err)
	}
	if loginResponse.StatusCode == apischema.STATUS_CODE_DATA {
		// successful login
		fmt.Println(color.GreenString("Successfully logged in"))
		config.Config.Token = loginResponse.Data.Id
		config.Config.Email = email
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
