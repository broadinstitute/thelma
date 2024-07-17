package terra

import "time"

type Environment interface {
	// DefaultCluster Returns the default cluster for this environment.
	DefaultCluster() Cluster
	// Namespace Returns the namespace for this environment. Eg. "terra-dev"
	Namespace() string
	// Lifecycle returns the lifecycle for this environment.
	Lifecycle() Lifecycle
	// Template returns the name of this environment's configuration template, if it has one.
	// Returns the empty string if the environment has no configuration template.
	Template() string
	// BaseDomain returns static domain name part for this environment or environment type.
	// E.g. "bee.envs-terra.bio", "dsde-prod.broadinstitute.org"
	BaseDomain() string
	// NamePrefixesDomain returns whether this particular environment's name should come before its BaseDomain when
	// deriving full hostnames/URLs in this environment.
	// E.g. 'true' for dynamic/template environments, 'false' for static
	NamePrefixesDomain() bool
	// UniqueResourcePrefix (dynamic environments only) unique-to-this-environment 4-character prefix that can be referenced in configuration.
	// Format: [a-z][a-z0-9]{3}
	// Returns empty string for static / template environments
	UniqueResourcePrefix() string
	// Owner is an email address of the user or group responsible for this environment.
	// May be empty if there's no owner or if the state provider doesn't track this information.
	Owner() string
	// PreventDeletion if true, the environment should not be automatically deleted under any circumstances.
	// Applies to dynamic environments only (Thelma only supports deletion of dynamic environments).
	PreventDeletion() bool
	// AutoDelete automatic deletion settings for this environment. Applies to dynamic environments only.
	AutoDelete() AutoDelete
	// CreatedAt returns the timestamp at which this environment was created in state
	CreatedAt() time.Time
	// Offline returns whether this environment should be currently offline
	Offline() bool
	// OfflineScheduleBeginEnabled indicates whether the environment is meant to be stopped on a schedule
	OfflineScheduleBeginEnabled() bool
	OfflineScheduleBeginTime() time.Time
	// OfflineScheduleEndEnabled indicates whether the environment is meant to be started on a schedule
	OfflineScheduleEndEnabled() bool
	OfflineScheduleEndTime() time.Time
	// OfflineScheduleEndWeekends indicates whether the start schedule should only apply on weekdays
	OfflineScheduleEndWeekends() bool
	// EnableJanitor indicates whether the Janitor service should be used for this environment to help reduce cloud costs.
	EnableJanitor() bool

	Destination
}
