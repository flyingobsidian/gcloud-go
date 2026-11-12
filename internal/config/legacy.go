package config

// The functions in this file use the process environment. They let callers
// that haven't moved to an explicit Env keep working while command groups are
// migrated (#2070), and will be removed once nothing uses them.

// ConfigDir returns the gcloud configuration directory. See Env.ConfigDir.
func ConfigDir() (string, error) { return OSEnv().ConfigDir() }

// Load reads the active configuration. See Env.Load.
func Load() (*Properties, error) { return OSEnv().Load() }

// LoadNamed reads a named configuration. See Env.LoadNamed.
func LoadNamed(name string) (*Properties, error) { return OSEnv().LoadNamed(name) }

// Save writes p to the active configuration. See Env.Save.
func (p *Properties) Save() error { return OSEnv().Save(p) }

// ActiveConfigName returns the active configuration name. See Env.ActiveConfigName.
func ActiveConfigName() string { return OSEnv().ActiveConfigName() }

// CreateConfiguration creates a named configuration. See Env.CreateConfiguration.
func CreateConfiguration(name string) error { return OSEnv().CreateConfiguration(name) }

// ActivateConfiguration activates a named configuration. See Env.ActivateConfiguration.
func ActivateConfiguration(name string) error { return OSEnv().ActivateConfiguration(name) }

// ListConfigurations lists configuration names. See Env.ListConfigurations.
func ListConfigurations() ([]string, error) { return OSEnv().ListConfigurations() }

// DeleteConfiguration removes a named configuration. See Env.DeleteConfiguration.
func DeleteConfiguration(name string) error { return OSEnv().DeleteConfiguration(name) }

// Resolve returns flag, else $envVar, else configVal. See Env.Resolve.
func Resolve(flag, envVar, configVal string) string {
	return OSEnv().Resolve(flag, envVar, configVal)
}

// Property returns section/key from the environment or active configuration.
// See Env.Property.
func Property(section, key string) (string, error) { return OSEnv().Property(section, key) }

// SetProperty sets section/key in the active configuration. See Env.SetProperty.
func SetProperty(section, key, value string) error {
	return OSEnv().SetProperty(section, key, value)
}

// UnsetProperty removes section/key from the active configuration. See Env.UnsetProperty.
func UnsetProperty(section, key string) error { return OSEnv().UnsetProperty(section, key) }

// FileProperties returns every property in the active configuration file.
// See Env.FileProperties.
func FileProperties() (map[string]map[string]string, error) { return OSEnv().FileProperties() }

// LoadAdditionalRepositories returns the persisted component repositories.
// See Env.LoadAdditionalRepositories.
func LoadAdditionalRepositories() ([]string, error) { return OSEnv().LoadAdditionalRepositories() }

// SaveAdditionalRepositories persists the component repositories.
// See Env.SaveAdditionalRepositories.
func SaveAdditionalRepositories(repos []string) error {
	return OSEnv().SaveAdditionalRepositories(repos)
}
