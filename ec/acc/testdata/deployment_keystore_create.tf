data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "%s"
}

resource "ec_deployment" "test" {
  name                   = "%s"
  region                 = "%s"
  version                = data.ec_stack.latest.version
  deployment_template_id = "%s"

  elasticsearch = {
    hot = {
      size        = "1g"
      autoscaling = {}
    }

    ml = {
      autoscaling = {}
    }

    config = {
      user_settings_yaml = <<EOF
xpack.security.authc.realms.oidc.oidc1:
  order: 2
  rp.client_id: 0oa5mogiwihtQzPzn697
  rp.response_type: "code"
  rp.requested_scopes: ["openid", "email"]
  rp.redirect_uri: "https://es-test.192.168.44.10.ip.es.io:9243/api/security/oidc/callback"
  op.issuer: "https://keystore-test.okta.com"
  op.authorization_endpoint: "https://keystore-test.okta.com/oauth2/v1/authorize"
  op.token_endpoint: "https://keystore-test.okta.com/oauth2/v1/token"
  op.userinfo_endpoint: "https://keystore-test.okta.com/oauth2/v1/userinfo"
  op.endsession_endpoint: "https://keystore-test.okta.com/oauth2/v1/logout"
  op.jwkset_path: "https://keystore-test.okta.com/oauth2/v1/keys"
  claims.principal: email
  claim_patterns.principal: "^([^@]+)@elasticsearch\\.com$"
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
    description: "Log in with Okta - test"
  basic.basic1:
    order: 1
EOF
    }
  }
}
