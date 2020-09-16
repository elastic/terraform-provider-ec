.PHONY: clean
clean: ## Delete credentials
	@ rm -f .apikey

.apikey:
	@ vault read -field=apikey secret/cloud-team/production/delivery-clients > .apikey
