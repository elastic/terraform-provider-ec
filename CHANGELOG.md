# 0.10.0 (Unreleased)

FEATURES:

* datasource/deployments: Adds additional parameter `name` to allow searching by exact deployment name. ([#797](https://github.com/elastic/terraform-provider-ec/issues/797))
* datasource/deploymenttemplates: Adds a new datasource to list all deployment-templates available in a region. ([#799](https://github.com/elastic/terraform-provider-ec/issues/799))
* resource/deployment: Added support for autoscaling Machine Learning tier only ([#761](https://github.com/elastic/terraform-provider-ec/issues/761))
* resource/deployment: Added support for symbols and profiling endpoints. ([#783](https://github.com/elastic/terraform-provider-ec/issues/783))
* resource/deployment: Validate the Kibana is present when attempting to enable other stateless resources. ([#792](https://github.com/elastic/terraform-provider-ec/issues/792))

ENHANCEMENTS:

* provider: Remove direct dependency on the old Terraform Plugin SDK ([#720](https://github.com/elastic/terraform-provider-ec/issues/720))
* provider: Update go version to 1.21 ([#713](https://github.com/elastic/terraform-provider-ec/issues/713))
* resource/deployment: Add support for instance configuration versions
  * Add instance_configuration_version field to all resources and allow to update the instance_configuration_id to a value not defined in the template.
  * Add migrate_to_latest_hardware field to allow migrating to the latest deployment template values.
  * Add latest_instance_configuration_id and latest_instance_configuration_version read-only fields. ([#755](https://github.com/elastic/terraform-provider-ec/issues/755))

BUG FIXES:

* resource/deployment: Don't rewrite the observability deployment ID to `self` when it's been explicitly configured. ([#789](https://github.com/elastic/terraform-provider-ec/issues/789))
* resource/deployment: Fix issue setting the elasticsearch_username when resetting the elasticsearch_password ([#777](https://github.com/elastic/terraform-provider-ec/issues/777))
* resource/deployment: Fix segfaults during Create/Update
  * When `elasticsearch` attribute contains both `strategy` and `snapshot_source`.
  * When `elasticsearch` defines `snapshot` with `repository` that doesn't contain `reference`. ([#719](https://github.com/elastic/terraform-provider-ec/issues/719))
* resource/deployment: Persist the snapshot source settings during reads. This fixes a [provider crash](https://github.com/elastic/terraform-provider-ec/issues/787) when creating a deployment from a snapshot. ([#788](https://github.com/elastic/terraform-provider-ec/issues/788))
* resource/deployment: Update the elasticsearch_username when resetting the password. ([#752](https://github.com/elastic/terraform-provider-ec/issues/752))
* resource/extension: Fix provider crash when updating the contents of an extension. ([#749](https://github.com/elastic/terraform-provider-ec/issues/749))

# 0.9.0 (September 22, 2023)

FEATURES:

* resource/deployment: new "elasticsearch"'s "keystore_contents" attribute to manage deployment keystore items during deployment create and update calls. ([#674](https://github.com/elastic/terraform-provider-ec/issues/674))

ENHANCEMENTS:

* resource/deployment: Set the deployment ID in state as soon as possible to avoid an unmanaged deployment as a result of a subsequent failure. ([#690](https://github.com/elastic/terraform-provider-ec/issues/690))
* resource/deployment: Validates that the node_types/node_roles configuration used is supported by the specified Stack version. ([#683](https://github.com/elastic/terraform-provider-ec/issues/683))

BUG FIXES:

* datasource/deployment: Prevent a provider crash when the deployment data source is referencing a deleted deployment ([#688](https://github.com/elastic/terraform-provider-ec/issues/688))
* resource/deployment: Prevent an endless diff loop after importing deployments with APM or Integrations Server resources. ([#689](https://github.com/elastic/terraform-provider-ec/issues/689))
* resource/deployment: Prevent endless diff loops when deployment trust settings are empty ([#687](https://github.com/elastic/terraform-provider-ec/issues/687))

# 0.8.0 (August 5, 2023)

FEATURES:

* Upgrades the provider to terraform-plugin-framework:1.2.0 ([#660](https://github.com/elastic/terraform-provider-ec/issues/660))
* datasource/privatelink: Adds data sources (`aws_privatelink_endpoint`, `azure_privatelink_endpoint`, and `gcp_private_service_connect_endpoint`) to lookup private networking endpoint information. ([#659](https://github.com/elastic/terraform-provider-ec/issues/659))
* resource/deployment: Add `reset_elasticsearch_password` attribute to the deployment resource. When true, this will reset the system password for the target deployment, updating the `elasticsearch_password` output as a result. ([#642](https://github.com/elastic/terraform-provider-ec/issues/642))
* resource/deployment: Adds endpoints integrations server resources. This allows consumers to explicitly capture service urls for dependent modules (e.g APM and Fleet). ([#640](https://github.com/elastic/terraform-provider-ec/issues/640))
* Prevents traffic filters managed with the `ec_deployment_traffic_filter_association` from being disassociated by the `ec_deployment` resource ([#419](https://github.com/elastic/terraform-provider-ec/issues/419)). This also fixes a provider crash for the above scenario present in 0.6 ([#621](https://github.com/elastic/terraform-provider-ec/issues/621)) ([#632](https://github.com/elastic/terraform-provider-ec/issues/632))
* resource/deployment: Fix validation and application of elasticsearch plan strategy. ([#648](https://github.com/elastic/terraform-provider-ec/issues/648))
* resource/deployment: Fix a value conversion error encountered when attempting to parse deployments without a snapshot repository. ([#666](https://github.com/elastic/terraform-provider-ec/issues/666))
* datasource/deployments: Fix bug causing a provider crash when no autoscaling fields are defined in the matching deployment. ([#667](https://github.com/elastic/terraform-provider-ec/issues/667))
* provider: Fix incompatibilities causing infinite configuration drift when used with Terraform CLI 1.4 or higher. ([#677](https://github.com/elastic/terraform-provider-ec/issues/677))
* resource/deployment: Fix bugs related to transitioning to/from deployment topologies which include dedicated master nodes. ([#682](https://github.com/elastic/terraform-provider-ec/issues/682))

# 0.7.0 (May 4, 2023)

ENHANCEMENTS:

* Add resource ec_snapshot_repository for usage with Elastic Cloud Enterprise. ([#613](https://github.com/elastic/terraform-provider-ec/issues/613))
* data-source/traffic_filter: Add `ec_traffic_filter` data source. ([#619](https://github.com/elastic/terraform-provider-ec/issues/619))
* resource/deployment: Ignore stopped resources when calculating the deployment version. ([#623](https://github.com/elastic/terraform-provider-ec/issues/623))
* resource/ec_deployment: Add snapshot settings (for usage with Elastic Cloud Enterprise only). ([#620](https://github.com/elastic/terraform-provider-ec/issues/620))
* resource/ec_deployment: Support the template migration api when changing deployment_template_id. ([#625](https://github.com/elastic/terraform-provider-ec/issues/625))

# 0.6.0 (Feb 28, 2023)

FEATURES:

Migration to [TF Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework)

**BREAKING CHANGES**:

New schema for `ec_deployment`. Existing resources should be imported. Please see NOTES below and README for more details.

BUG FIXES:

[#336](https://github.com/elastic/terraform-provider-ec/issues/336)
[#445](https://github.com/elastic/terraform-provider-ec/issues/445)
[#466 ](https://github.com/elastic/terraform-provider-ec/issues/466)
[#467](https://github.com/elastic/terraform-provider-ec/issues/467)

NOTES

* Older versions of terraform CLI can report errors with the provider 0.6.0. Please make sure to update Terraform CLI to the latest version.
* `ec_deployment` has a new schema now but state upgrade is not implemented.
  The recommended way to proceed with existing TF resources is [state import](https://developer.hashicorp.com/terraform/cli/import#state-only).
  However, this doesn't import user passwords and secret tokens.
* After import, the next plan command may try to delete some empty or zero size attributes, e.g. it can try to delete empty `elasticsearch` `config` or `cold` tier if configuration doesn't define them and `cold` tier size is zero.
  It should not be a problem. You can eigher execute the plan (the only result should be updated Terraform state while the deployment should stay the same) or add empty `cold` tier and `config` attributes to the configuration.
* The migration is based on 0.4.1, so all changes from 0.5.0 and 0.5.1 are omitted.

# 0.5.1 (Feb 15, 2023)

FEATURES:

* resource/deployment: Utilise the template migration API to build the base update request when changing `deployment_template_id`. This results in more reliable changes between deployment templates. ([#547](https://github.com/elastic/terraform-provider-ec/issues/547))

# 0.5.0 (Oct 12, 2022)

FEATURES:

* datasource/privatelink: Adds data sources to obtain AWS/Azure Private Link, and GCP Private Service Connect configuration data. ([#533](https://github.com/elastic/terraform-provider-ec/issues/533))
* resource/deployment: Adds fleet_https_endpoint and apm_https_endpoint to integrations server resources. This allows consumers to explicitly capture service urls for dependent modules. ([#548](https://github.com/elastic/terraform-provider-ec/issues/548))
* resource/elasticsearch: Adds support for the `strategy` property to the `elasticsearch` resource. This allows users to define how different plan changes are coordinated. ([#507](https://github.com/elastic/terraform-provider-ec/issues/507))

BUG FIXES:

* resource/deployment: Correctly restrict stateless (Kibana/Enterprise Search/Integrations Server) resources to a single topology element. Fixes a provider crash when multiple elements without an instance_configuration_id were specified. ([#536](https://github.com/elastic/terraform-provider-ec/issues/536))
* resource/elasticsearchkeystore: Correctly delete keystore items when removed from the module definition. ([#546](https://github.com/elastic/terraform-provider-ec/issues/546))
* resource: Updates all nested field accesses to validate type casts. This prevents a provider crash when a field is explicitly set to `null`. ([#534](https://github.com/elastic/terraform-provider-ec/issues/534))

# 0.4.1 (May 11, 2022)

BREAKING CHANGES:

* To support unsized topology elements when autoscaling is enabled, we now include all potentially sized topology elements in the `ec_deployment` state.
When autoscaling is enabled, we now require that all autoscaleable topology elements be defined in the `elasticsearch` block of an `ec_deployment` resource.
If a topology element is not defined, Terraform will report a persistent diff during a plan/apply. ([#472](https://github.com/elastic/terraform-provider-ec/issues/472))

BUG FIXES:

* Allow zero sized topology elements when autoscaling is enabled. Previously, including an ML topology block would result in a persistent diff loop when the underlying ML tier remained disabled by autoscaling (i.e no ML jobs were enabled). ([#472](https://github.com/elastic/terraform-provider-ec/issues/472))
* main: Adds debug mode. Instructions for debugging the provider can be found in the [CONTRIBUTING](https://github.com/elastic/terraform-provider-ec/blob/master/CONTRIBUTING.md#debugging) docs. ([#430](https://github.com/elastic/terraform-provider-ec/issues/430))

# 0.4.0 (Feb 24, 2022)

FEATURES:

* resource/integrations_server: Adds a new `ec_deployment_integrations_server` resource to the deployment, which has been introduced in Elastic Stack 8.0.0 ([#425](https://github.com/elastic/terraform-provider-ec/issues/425))

# 0.3.0 (Oct 17, 2021)

FEATURES:

* **New Resource:** resource/ec_deployment_elasticsearch_keystore: Adds a new `ec_deployment_elasticsearch_keystore` resource which allows creating and updating Elasticsearch keystore settings. ([#364](https://github.com/elastic/terraform-provider-ec/issues/364))

ENHANCEMENTS:

* datasource/ec_deployments: Adds four new fields, `deployments.#.elasticsearch_ref_id`, `deployments.#.kibana_ref_id`, `deployments.#.apm_ref_id`, `deployments.#.enterprise_search_ref_id` to the data source. ([#380](https://github.com/elastic/terraform-provider-ec/issues/380))
* datasource/ec_deployments: Adds two new fields, `deployments.#.name` and `deployments.#.alias` to the data source. ([#362](https://github.com/elastic/terraform-provider-ec/issues/362))
* resource/ec_deployment_traffic_filter: Add support for Azure Private Link traffic rules. ([#340](https://github.com/elastic/terraform-provider-ec/issues/340))

BUG FIXES:

* resource/ec_deployment: Changes the `ec_deployment.elasticsearch.remote_cluster` block to `schema.TypeSet` to allow specifying the blocks in any order. ([#368](https://github.com/elastic/terraform-provider-ec/issues/368))
* resource/ec_deployment: Fix bug where setting any of the `elasticsearch.config.user_settings_* = null` would result in a provider panic. ([#355](https://github.com/elastic/terraform-provider-ec/issues/355))
* resource/ec_deployment: Fix bug where some of the settings that were set by the UI were unset by the Terraform provider. See #214 for more details on the bug report. ([#361](https://github.com/elastic/terraform-provider-ec/issues/361))
* resource/ec_deployment: Fix bug where the deployment alias is ignored. ([#341](https://github.com/elastic/terraform-provider-ec/issues/341))
* resource/ec_deployment: Fixed a bug that affects partial version upgrades. During an upgrade only a subset of resources would upgrade successfully, but the `version` argument value updated as if all resources were upgraded. Attempts to retry the upgrade would fail since the version difference was not detected. ([#371](https://github.com/elastic/terraform-provider-ec/issues/371))

# 0.2.1 (Jun 17, 2021)

BUG FIXES:

* resource/ec_deployment: Fixes a bug which made ec_deployment version upgrades return an API error stating: `node_roles must be provided for all elasticsearch topology elements or for none of them`. ([#329](https://github.com/elastic/terraform-provider-ec/issues/329))

# 0.2.0 (Jun 15, 2021)

FEATURES:

* datasource/ec_deployment: Add a new size parameter to allow modifying the default size of `10` in the number of deployments returned by the search request. ([#300](https://github.com/elastic/terraform-provider-ec/issues/300))
* resource/ec_deployment: Supports Autoscaling via two new settings: `elasticsearch.autoscale` (`"true"` or `"false"`) and an `elasticsearch.topology.autoscaling` block to modify the default autoscaling policies. For more information, refer to the [documentation examples](https://registry.terraform.io/providers/elastic/ec/latest/docs/resources/ec_deployment#example-usage). ([#296](https://github.com/elastic/terraform-provider-ec/issues/296))
* resource/ec_deployment: Supports deployment aliases in a new top level field `alias`. ([#298](https://github.com/elastic/terraform-provider-ec/issues/298))

ENHANCEMENTS:

* resource/ec_deployment: Retries the Shutdown API call on the destroy operation up to 3 times when the transient "Timeout Exceeded" error returned from the Elastic Cloud API. ([#308](https://github.com/elastic/terraform-provider-ec/issues/308))

BUG FIXES:

* datasource/ec_deployments: Properly sorts the datasource results by ID. ([#322](https://github.com/elastic/terraform-provider-ec/issues/322))
* resource/ec_deployment: Fixes a bug which made restoring a snapshot to an existing deployment fail. ([#309](https://github.com/elastic/terraform-provider-ec/issues/309))
* resource/ec_deployment: Handles account and external trust settings, fixing a bug where the default trust settings are unset and allowing users to set up their own trust settings for an Elasticsearch cluster. ([#324](https://github.com/elastic/terraform-provider-ec/issues/324))

# 0.1.1 (April 7, 2021)

BUG FIXES:

* resource/ec_deployment: Fixes a bug where specifying a dedicated tier for master or coordinating nodes would result in an API stating that the master or ingest roles are duplicated. ([#291](https://github.com/elastic/terraform-provider-ec/issues/291))

# 0.1.0 (March 31, 2021)

BREAKING CHANGES:

* ec_deployment: Removes the `apm.version`, `enterprise_search.version` and `kibana.version` computed fields. ([#266](https://github.com/elastic/terraform-provider-ec/issues/266))
* resource/ec_deployment: Adds support for the newly added data tiers. A new **required** field `elasticsearch.toplogy.id` has been added, it needs to be set to all **explicit** Elasticsearch topology declarations. A `node_roles` computed field has been added to the schema and **cannot** be overridden by the user, versions `>=7.10.0` will be automatically migrated by the provider to use `node_roles` from the `node_type_*` settings, these will be removed from the state. When `node_type_*` fields are explicitly set in the terraform configuration they need to be unset manually by the user. Additionally, it removes the `elasticsearch.version` computed field. ([#253](https://github.com/elastic/terraform-provider-ec/issues/253))

FEATURES:

* **New Resource:** resource/ec_extension: Add a new resource `ec_extension` which allows users to mange custom Elasticsearch bundles and plugins ([#216](https://github.com/elastic/terraform-provider-ec/issues/216))

ENHANCEMENTS:

* datasource/ec_deployment: Adds the tag attribute to the `ec_deployment` datasource ([#244](https://github.com/elastic/terraform-provider-ec/issues/244))
* datasource/ec_deployments: Allows filtering deployments by their associated tags ([#248](https://github.com/elastic/terraform-provider-ec/issues/248))
* resource/ec_deployment: Add tags key / value map ([#218](https://github.com/elastic/terraform-provider-ec/issues/218))
* resource/ec_deployment: Adds a new `elasticsearch.extension` block which can be used to enable custom Elasticsearch bundles or plugins that have previously been uploaded. ([#264](https://github.com/elastic/terraform-provider-ec/issues/264))

BUG FIXES:

* datasource/ec_deployment: Fixes bug where the datasource was persisting zero sized topology elements in the state ([#242](https://github.com/elastic/terraform-provider-ec/issues/242))
* datasource/ec_deployments: Fixes bug where queries containing a hyphens wouldn't work as expected ([#241](https://github.com/elastic/terraform-provider-ec/issues/241))
* go/build: Fixes bug where the api user agent wasn't stripped of its `-dev` tag prior to releasing ([#235](https://github.com/elastic/terraform-provider-ec/issues/235))
* resource/ec_traffic_filter: Fixes bug where having a traffic filter with a multiple rules will cause an infinite diff due to ordering ([#208](https://github.com/elastic/terraform-provider-ec/issues/208))

# 0.1.0-beta (December 14, 2020)

NOTES

The Elastic Cloud Terraform provider allows you to provision Elastic Cloud deployments on any Elastic Cloud platform, whether it’s Elasticsearch Service or Elastic Cloud Enterprise.

_This functionality is in beta and is subject to change. The design and code are less mature than official GA features and are being provided as-is with no warranties._

FEATURES

* **New Provider**: ec ([docs](https://registry.terraform.io/providers/elastic/ec/0.1.0-beta/docs))
* **New Resource**: ec_deployment ([docs](https://registry.terraform.io/providers/elastic/ec/latest/docs/resources/ec_deployment))
* **New Resource**: ec_deployment_traffic_filter ([docs](https://registry.terraform.io/providers/elastic/ec/latest/docs/resources/ec_deployment_traffic_filter))
* **New Resource**: ec_deployment_traffic_filter_association ([docs](https://registry.terraform.io/providers/elastic/ec/latest/docs/resources/ec_deployment_traffic_filter_association))
* **New Data Source**: ec_deployment ([docs](https://registry.terraform.io/providers/elastic/ec/latest/docs/data-sources/ec_deployment))
* **New Data Source**: ec_deployments ([docs](https://registry.terraform.io/providers/elastic/ec/latest/docs/data-sources/ec_deployments))
* **New Data Source**: ec_stack ([docs](https://registry.terraform.io/providers/elastic/ec/latest/docs/data-sources/ec_stack))
