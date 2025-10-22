# execute a release after merging a PR to main
# example: `make release VERSION=v1.0.3`
.PHONY: release
release:
	@if [ -z "$(VERSION)" ]; then \
		echo "❌ VERSION required. Usage: make release VERSION=v1.0.3"; \
		exit 1; \
	fi; \
	echo "Creating release $(VERSION)..."; \
	git tag $(VERSION); \
	git push origin $(VERSION); \
	$(MAKE) sync-v1

.PHONY: sync-v1
sync-v1:
	@echo "Syncing v1 branch with latest v1.* tag..."
	@latest_tag=$$(git tag -l "v1.*" --sort=-creatordate | head -n 1); \
	if [ -z "$$latest_tag" ]; then \
		echo "❌ No v1.* tags found"; \
		exit 1; \
	fi; \
	echo "Latest tag: $$latest_tag"; \
	git fetch --all --tags; \
	git checkout tags/$$latest_tag -b tmp-v1; \
	git push origin tmp-v1:v1 --force; \
	git checkout main; \
	git branch -D tmp-v1; \
	echo "v1 branch updated to $$latest_tag"