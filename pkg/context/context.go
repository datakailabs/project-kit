package context

import (
	"fmt"
	"os/exec"

	"github.com/datakaicr/pk/pkg/config"
)

// Switch switches cloud and git contexts based on project configuration
func Switch(project *config.Project) error {
	if project.Context.GitIdentity == "" &&
		project.Context.AWSProfile == "" &&
		project.Context.AzureSubscription == "" &&
		project.Context.GCloudProject == "" &&
		project.Context.DatabricksProfile == "" &&
		project.Context.SnowflakeAccount == "" {
		// No context configured
		return nil
	}

	fmt.Printf("☁️  Switching context for: %s\n", project.ProjectInfo.Name)

	// Switch git identity
	if project.Context.GitIdentity != "" {
		if err := switchGitIdentity(project.Context.GitIdentity); err != nil {
			fmt.Printf("Warning: Failed to switch git identity: %v\n", err)
		} else {
			fmt.Printf("   Git: %s\n", project.Context.GitIdentity)
		}
	}

	// Switch AWS profile
	if project.Context.AWSProfile != "" {
		if err := switchAWSProfile(project.Context.AWSProfile); err != nil {
			fmt.Printf("Warning: Failed to switch AWS profile: %v\n", err)
		} else {
			fmt.Printf("   AWS: %s\n", project.Context.AWSProfile)
		}
	}

	// Switch Azure subscription
	if project.Context.AzureSubscription != "" {
		if err := switchAzureSubscription(project.Context.AzureSubscription); err != nil {
			fmt.Printf("Warning: Failed to switch Azure subscription: %v\n", err)
		} else {
			fmt.Printf("   Azure: %s\n", project.Context.AzureSubscription)
		}
	}

	// Switch GCloud project
	if project.Context.GCloudProject != "" {
		if err := switchGCloudProject(project.Context.GCloudProject); err != nil {
			fmt.Printf("Warning: Failed to switch GCloud project: %v\n", err)
		} else {
			fmt.Printf("   GCloud: %s\n", project.Context.GCloudProject)
		}
	}

	// Switch Databricks profile
	if project.Context.DatabricksProfile != "" {
		fmt.Printf("   Databricks: %s\n", project.Context.DatabricksProfile)
		// Databricks uses env var, set in session
	}

	// Switch Snowflake account
	if project.Context.SnowflakeAccount != "" {
		fmt.Printf("   Snowflake: %s\n", project.Context.SnowflakeAccount)
		// Snowflake uses env var, set in session
	}

	return nil
}

func switchGitIdentity(identity string) error {
	// Check if git is installed
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git not installed")
	}

	// This sets local git config for the project directory
	// The actual switching happens in the project directory
	return nil
}

func switchAWSProfile(profile string) error {
	// Check if aws CLI is installed
	if _, err := exec.LookPath("aws"); err != nil {
		return fmt.Errorf("aws CLI not installed")
	}

	// Export AWS_PROFILE environment variable (done in session)
	return nil
}

func switchAzureSubscription(subscription string) error {
	// Check if az CLI is installed
	if _, err := exec.LookPath("az"); err != nil {
		return fmt.Errorf("azure CLI not installed")
	}

	// Set default subscription
	cmd := exec.Command("az", "account", "set", "--subscription", subscription)
	return cmd.Run()
}

func switchGCloudProject(project string) error {
	// Check if gcloud is installed
	if _, err := exec.LookPath("gcloud"); err != nil {
		return fmt.Errorf("gcloud CLI not installed")
	}

	// Set default project
	cmd := exec.Command("gcloud", "config", "set", "project", project)
	return cmd.Run()
}
