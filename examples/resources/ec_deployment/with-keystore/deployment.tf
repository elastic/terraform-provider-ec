# Retrieve the latest stack pack version
data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "us-east-1"
}

# Create an Elastic Cloud deployment with keystore
resource "ec_deployment" "with_keystore" {
  name = "example_with_keystore"

  region                 = "us-east-1"
  version                = data.ec_stack.latest.version
  deployment_template_id = "aws-io-optimized-v2"

  elasticsearch = {
    hot = {
      autoscaling = {}
    }

    config = {
      user_settings_yaml = <<EOF
xpack.security.authc.realms.oidc.oidc1:
  order: 1
  rp.client_id: "<client-id>"
  rp.response_type: "code"
  rp.requested_scopes: ["openid", "email"]
  rp.redirect_uri: "<KIBANA_ENDPOINT_URL>/api/security/oidc/callback"
  op.issuer: "<YOUR_OKTA_DOMAIN>"
  op.authorization_endpoint: "<YOUR_OKTA_DOMAIN>/oauth2/v1/authorize"
  op.token_endpoint: "<YOUR_OKTA_DOMAIN>/oauth2/v1/token"
  op.userinfo_endpoint: "<YOUR_OKTA_DOMAIN>/oauth2/v1/userinfo"
  op.endsession_endpoint: "<YOUR_OKTA_DOMAIN>/oauth2/v1/logout"
  op.jwkset_path: "<YOUR_OKTA_DOMAIN>/oauth2/v1/keys"
  claims.principal: email
  claim_patterns.principal: "^([^@]+)@elastic\\.co$"
EOF
    }

    keystore_contents = {
      "xpack.security.authc.realms.oidc.oidc1.rp.client_secret" = {
         value = "secret-1"
      }
    }
  }

  kibana = {
    zone_count = 1
    config = {
      user_settings_yaml = <<EOF
xpack.security.authc.providers:
  oidc.oidc1:
    order: 0
    realm: oidc1
    description: "Log in with Okta"
  basic.basic1:
    order: 1
EOF
    }
  }
}
