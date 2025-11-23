package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Project represents a .project.toml file
type Project struct {
	Path string // Full path to project directory

	// ==========================================
	// CORE SCHEMA (universal, always present)
	// ==========================================

	// [project] section
	ProjectInfo struct {
		Name   string `toml:"name"`
		ID     string `toml:"id"`
		Status string `toml:"status"`
		Type   string `toml:"type"`
	} `toml:"project"`

	// [tech] section
	Tech struct {
		Stack  []string `toml:"stack"`
		Domain []string `toml:"domain"`
	} `toml:"tech"`

	// [dates] section
	Dates struct {
		Started   string `toml:"started"`
		Completed string `toml:"completed"`
	} `toml:"dates"`

	// [links] section (core - generic links only)
	Links struct {
		Repository         string `toml:"repository"`
		Documentation      string `toml:"documentation"`
		ScriptoriumProject string `toml:"scriptorium_project"` // LEGACY - migrates to datakai
		ConduitGraph       string `toml:"conduit_graph"`       // LEGACY - migrates to datakai
	} `toml:"links"`

	// [notes] section
	Notes struct {
		Description string `toml:"description"`
	} `toml:"notes"`

	// [tmux] section (optional)
	Tmux struct {
		Layout  string       `toml:"layout"`
		Windows []TmuxWindow `toml:"windows"`
	} `toml:"tmux"`

	// [context] section (optional)
	Context struct {
		AWSProfile        string `toml:"aws_profile"`
		AzureSubscription string `toml:"azure_subscription"`
		GCloudProject     string `toml:"gcloud_project"`
		DatabricksProfile string `toml:"databricks_profile"`
		SnowflakeAccount  string `toml:"snowflake_account"`
		GitIdentity       string `toml:"git_identity"`
	} `toml:"context"`

	// ==========================================
	// CONSULTANT EXTENSION (optional)
	// ==========================================

	Consultant struct {
		Ownership       string `toml:"ownership"`        // datakai | client | shared | open-source
		ClientName      string `toml:"client_name"`      // "Acme Corp"
		ClientType      string `toml:"client_type"`      // direct | partner | internal
		Partner         string `toml:"partner"`          // "West Monroe"
		MyRole          string `toml:"my_role"`          // lead | contributor | advisor
		DeliverableType string `toml:"deliverable_type"` // product | consulting | support
		LicenseModel    string `toml:"license_model"`    // proprietary | client-owned | open-source
		Billable        bool   `toml:"billable"`
		RateType        string `toml:"rate_type"` // fixed | hourly | retainer
	} `toml:"consultant"`

	// ==========================================
	// DATAKAI EXTENSION (optional)
	// ==========================================

	DataKai struct {
		Visibility         string   `toml:"visibility"` // CRITICAL: private | public | client-confidential
		ScriptoriumProject string   `toml:"scriptorium_project"`
		ConduitGraph       string   `toml:"conduit_graph"`
		Protocols          []string `toml:"protocols"`
		DKOSVersion        string   `toml:"dkos_version"`
		ProductCategory    string   `toml:"product_category"` // infrastructure | client-deliverable | internal-tool
		RevenueModel       string   `toml:"revenue_model"`    // saas | consulting | open-source | internal
		Maturity           string   `toml:"maturity"`         // experimental | mvp | production | deprecated
	} `toml:"datakai"`

	// ==========================================
	// LEGACY FIELDS (backward compatibility)
	// Auto-migrated on load, removed on save
	// ==========================================

	LegacyOwnership struct {
		Primary      string   `toml:"primary"`
		Partners     []string `toml:"partners"`
		LicenseModel string   `toml:"license_model"`
		Visibility   string   `toml:"visibility"` // WRONG location - migrates to datakai.visibility
	} `toml:"ownership"`

	LegacyClient struct {
		EndClient    string `toml:"end_client"`
		Intermediary string `toml:"intermediary"`
		MyRole       string `toml:"my_role"`
	} `toml:"client"`

	// Track if migration occurred
	migrated bool `toml:"-"`
}

// TmuxWindow represents a window configuration
type TmuxWindow struct {
	Name    string `toml:"name"`
	Command string `toml:"command"`
	Path    string `toml:"path"`
}

// LoadProject reads a .project.toml file
func LoadProject(path string) (*Project, error) {
	var project Project
	project.Path = filepath.Dir(path)

	// Decode TOML file
	if _, err := toml.DecodeFile(path, &project); err != nil {
		return nil, err
	}

	// Auto-migrate legacy schema to new format
	project.migrateSchema()

	return &project, nil
}

// GetOwner returns the project owner (backward compatibility)
func (p *Project) GetOwner() string {
	if p.Consultant.Ownership != "" {
		return p.Consultant.Ownership
	}
	return p.LegacyOwnership.Primary
}

// GetLicenseModel returns license model (backward compatibility)
func (p *Project) GetLicenseModel() string {
	if p.Consultant.LicenseModel != "" {
		return p.Consultant.LicenseModel
	}
	return p.LegacyOwnership.LicenseModel
}

// GetClientName returns client name (backward compatibility)
func (p *Project) GetClientName() string {
	if p.Consultant.ClientName != "" {
		return p.Consultant.ClientName
	}
	return p.LegacyClient.EndClient
}

// GetPartner returns partner name (backward compatibility)
func (p *Project) GetPartner() string {
	if p.Consultant.Partner != "" {
		return p.Consultant.Partner
	}
	if len(p.LegacyOwnership.Partners) > 0 {
		return p.LegacyOwnership.Partners[0]
	}
	return p.LegacyClient.Intermediary
}

// GetMyRole returns user's role (backward compatibility)
func (p *Project) GetMyRole() string {
	if p.Consultant.MyRole != "" {
		return p.Consultant.MyRole
	}
	return p.LegacyClient.MyRole
}

// GetPartners returns partner list (backward compatibility)
func (p *Project) GetPartners() []string {
	if p.Consultant.Partner != "" {
		return []string{p.Consultant.Partner}
	}
	return p.LegacyOwnership.Partners
}

// migrateSchema converts old schema format to new
func (p *Project) migrateSchema() {
	// Check if migration is needed
	hasLegacyOwnership := p.LegacyOwnership.Primary != ""
	hasLegacyClient := p.LegacyClient.EndClient != "" || p.LegacyClient.Intermediary != ""

	if !hasLegacyOwnership && !hasLegacyClient {
		// Still check for legacy links
		if p.Links.ScriptoriumProject == "" && p.Links.ConduitGraph == "" {
			return // No migration needed
		}
	}

	p.migrated = true

	// Migrate [ownership] -> [consultant]
	if hasLegacyOwnership {
		p.Consultant.Ownership = p.LegacyOwnership.Primary
		p.Consultant.LicenseModel = p.LegacyOwnership.LicenseModel

		// Partners array -> single partner field (take first)
		if len(p.LegacyOwnership.Partners) > 0 {
			p.Consultant.Partner = p.LegacyOwnership.Partners[0]
		}

		// CRITICAL: Migrate visibility from ownership to datakai
		if p.LegacyOwnership.Visibility != "" {
			p.DataKai.Visibility = p.LegacyOwnership.Visibility
		}
	}

	// Migrate [client] -> [consultant]
	if hasLegacyClient {
		p.Consultant.ClientName = p.LegacyClient.EndClient
		p.Consultant.MyRole = p.LegacyClient.MyRole

		// Intermediary -> partner + client_type
		if p.LegacyClient.Intermediary != "" {
			p.Consultant.Partner = p.LegacyClient.Intermediary
			p.Consultant.ClientType = "partner"
		} else if p.LegacyClient.EndClient != "" {
			p.Consultant.ClientType = "direct"
		}
	}

	// Migrate DataKai-specific links from [links] to [datakai]
	if p.Links.ScriptoriumProject != "" {
		p.DataKai.ScriptoriumProject = p.Links.ScriptoriumProject
		p.migrated = true
	}
	if p.Links.ConduitGraph != "" {
		p.DataKai.ConduitGraph = p.Links.ConduitGraph
		p.migrated = true
	}
}

// FindProjects recursively finds all .project.toml files
func FindProjects(rootDirs ...string) ([]*Project, error) {
	var projects []*Project

	for _, root := range rootDirs {
		// Check if directory exists
		if _, err := os.Stat(root); os.IsNotExist(err) {
			continue
		}

		// Walk directory tree
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Found a .project.toml file
			if info.Name() == ".project.toml" {
				project, err := LoadProject(path)
				if err != nil {
					// Skip malformed files
					return nil
				}
				projects = append(projects, project)
			}

			return nil
		})

		if err != nil {
			return nil, err
		}
	}

	return projects, nil
}
