ROOT_DIR               := $(CURDIR)/..

-include $(ROOT_DIR)/.env

vault :=$(ROOT_DIR)/scripts/retry.sh 5 vault

# BUILD_ID is present during run on Jenkins machine, but not on dev box, hence using it here to distinguish between those cases
ifndef VAULT_TOKEN
	ifdef BUILD_ID
		VAULT_TOKEN = $(shell $(vault) write -address=$(VAULT_ADDR) -field=token auth/approle/login role_id=$(VAULT_ROLE_ID) secret_id=$(VAULT_SECRET_ID))
	else
		VAULT_TOKEN = $(shell $(vault) write -address=$(VAULT_ADDR) -field=token auth/github/login token=$(GITHUB_TOKEN))
	endif
endif

.PHONY: clean
clean: ## Delete credentials
	@ rm -f .apikey

.apikey:
	@ $(vault) read -field=apikey secret/cloud-team/production/delivery-clients > .apikey
