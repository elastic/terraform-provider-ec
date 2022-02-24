# 0.4.0 (Unreleased)

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
