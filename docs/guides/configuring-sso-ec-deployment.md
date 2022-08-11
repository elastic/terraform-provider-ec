---
subcategory: ""
page_title: "Configuring an SSO provider for a deployment"
description: |-
    An example of how to create a deployment preconfigured with a SAML provider
---

# Configuring a SAML provider for an Elastic Cloud Deployment

A common use case for the Elastic Cloud Terraform provider i, is to spin up an elastic deployment preconfigured with an SSO Identity provider (SAML2.0 or OIDC based)

Since such a configuration often results in a cyclic dependency, where the Elasticsearch cluster yaml file requires the cluster full url name, but the url has not been generated yet, we're going to use Elastic Cloud's alias feature.

*Please note, this guide doesn't go through what's needed to be configured on the IDP side, as this can change for different identity providers*

First, we'll create an "random_uuid" resource, like so:

```terraform
resource "random_uuid" "uuid" {}
```

This will allow us to use a randomally generated uuid in our deployment name, so that we can later use it within the alias and revieve a unique deployment name & url alias, even before our deploymemnt is actually created. 

Creating the "ec_deployment" resource, will look like this:

```hcl
resource "ec_deployment" "elastic-sso" {
    name = format("%s-%s",var.name,substr("${random_uuid.uuid.result}",0,6))
    alias = format("%s-%s",var.name,substr("${random_uuid.uuid.result}",0,6))
    region = "us-east-1"
    version = "7.17.5"
    deployment_template_id = "aws-compute-optimized-v3"

    elasticsearch {
        topology {
            id = "hot_content"
            size = "8g"
            zone_count = 2
        }

        topology {
            id = "warm"
            size = "8g"
            zone_count = 2
        }

        config{
           user_settings_yaml = templatefile("./es.yml",{kibana_url=format("https://%s-%s.kb.us-east-1.aws.found.io:9243",var.name,substr("${random_uuid.uuid.result}",0,6))})
        }
    }

    kibana {
        config{
           user_settings_yaml = file("./kb.yml")
        }
    }
}

```

Let's take a closer look at one specific argument here:
```terraform
format("%s-%s",var.name,substr("${random_uuid.uuid.result}",0,6))
```

This will tell terraform to create a string, that looks like  ```<deployment-name>-<6-digits-of-uuid>``` 
We'll configure the deployment alias field to be the same, so if my deployment is named "elastic-sso" it'll get created as:
```elastic-sso-8f9f6s``` for example.

Then, using a variable in our es.yml file, and a terraform templating mechanism, we'll generate our proper es.yml file. Our variable is named kibana_url, as seen in the ec_deployment resource above: 

```terraform
    config{
           user_settings_yaml = templatefile("./es.yml",{kibana_url=format("https://%s-%s.kb.us-east-1.aws.found.io:9243",var.name,substr("${random_uuid.uuid.result}",0,6))})
        }
```

This specific template, will use our name and UUID to determine the url for our Elasticsearch deployment before its even created, and put it in our ```es.yml``` file.

```yaml
xpack.security.authc.realms.saml.auth0:
  order: 2
  idp.metadata.path: "https://<url-for-provider's metadata>"
  idp.entity_id: "urn:myproject.us.auth0.com"
  sp.entity_id:  "${kibana_url}/"
  sp.acs: "${kibana_url}/api/security/saml/callback"
  sp.logout: "${kibana_url}/logout"
  attributes.principal: "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/upn"
  attributes.groups: "http://schemas.xmlsoap.org/claims/Group"
```

Our kibana file should also contain the SSO provider configuration for Kibana to display it as a login option:

```yaml
xpack.security.authc.providers:
    saml.auth0:
        order: 1
        realm: auth0
        icon: logoSecurity
        description: "Log in with Auth0"
```

And that's it! Spinning up the above ec_deployment resource will create a deployment on Elastic Cloud, with a preconfigured name and an additional Auth0 SSO identity provider and login option, that's already configured when the deployment is spun up.