# start project configuration
name := aviation
buildDir := build
srcFiles := $(shell find . -name "*.go" -not -path "./$(buildDir)/*" -not -name "*_test.go" -not -path "*\#*")
packages := $(name) services
# end project configuration

# start environment setup
gobin := go
ifneq (,$(GOROOT))
gobin := $(GOROOT)/bin/go
endif

ifeq ($(OS),Windows_NT)
export GOCACHE := $(shell cygpath -m $(abspath $(buildDir)/.cache))
export GOLANGCI_LINT_CACHE := $(shell cygpath -m $(abspath $(buildDir)/.lint-cache))
export GOPATH := $(shell cygpath -m $(GOPATH))
export GOROOT := $(shell cygpath -m $(GOROOT))
endif

export GO111MODULE := off
# end environment setup

# Ensure the build directory exists, since most targets require it.
$(shell mkdir -p $(buildDir))

.DEFAULT_GOAL := compile

# start lint setup targets
lintDeps := $(buildDir)/golangci-lint $(buildDir)/.lintSetup $(buildDir)/run-linter
$(buildDir)/golangci-lint:
	@curl --retry 10 --retry-max-time 60 -sSfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(buildDir) v1.40.0 >/dev/null 2>&1 && touch $@
$(buildDir)/run-linter: cmd/run-linter/run-linter.go $(buildDir)/golangci-lint
	$(gobin) build -o $@ $<
# end lint setup targets

# start output files
testOutput := $(foreach target,$(packages),$(buildDir)/output.$(target).test)
lintOutput := $(foreach target,$(packages),$(buildDir)/output.$(target).lint)
coverageOutput := $(foreach target,$(packages),$(buildDir)/output.$(target).coverage)
coverageHtmlOutput := $(foreach target,$(packages),$(buildDir)/output.$(target).coverage.html)
.PRECIOUS: $(coverageOutput) $(coverageHtmlOutput) $(lintOutput) $(testOutput)
# end output files

# start basic development targets
compile: $(srcFiles)
	$(gobin) build $(subst $(name),,$(subst -,/,$(foreach target,$(packages),./$(target))))
test: $(testOutput)
lint: $(lintOutput)
coverage: $(coverageOutput)
coverage-html: $(coverageHtmlOutput)
phony += compile lint test coverage coverage-html

# start convenience targets for running tests and coverage tasks on a
# specific package.
test-%: $(buildDir)/output.%.test
	
coverage-%: $(buildDir)/output.%.coverage
	
html-coverage-%: $(buildDir)/output.%.coverage.html
	
lint-%: $(buildDir)/output.%.lint
	
# end convenience targets
# end basic development targets

# start test and coverage artifacts
testArgs := -v
ifneq (,$(RUN_TEST))
testArgs += -run='$(RUN_TEST)'
endif
ifneq (,$(RUN_COUNT))
testArgs += -count='$(RUN_COUNT)'
endif
ifneq (,$(SKIP_LONG))
testArgs += -short
endif
ifeq (,$(DISABLE_COVERAGE))
testArgs += -cover
endif
ifneq (,$(RACE_DETECTOR))
testArgs += -race
endif
$(buildDir)/output.%.test: .FORCE
	$(gobin) test $(testArgs) ./$(if $(subst $(name),,$*),$(subst -,/,$*),) | tee $@
	@! grep -s -q -e "^FAIL" $@ && ! grep -s -q "^WARNING: DATA RACE" $@
$(buildDir)/output.%.coverage: .FORCE
	$(gobin) test $(testArgs) -covermode=count -coverprofile ./$(if $(subst $(name),,$*),$(subst -,/,$*),) | tee $@
	@-[ -f $@ ] && $(gobin) tool cover -func=$@ | sed 's%$(projectPath)/%%' | column -t
$(buildDir)/output.%.coverage.html: $(buildDir)/output.%.coverage
	$(gobin) tool cover -html=$< -o $@

ifneq (go,$(gobin))
# We have to handle the PATH specially for linting in CI, because if the PATH has a different version of the Go
# binary in it, the linter won't work properly.
lintEnvVars := PATH="$(shell dirname $(gobin)):$(PATH)"
endif
$(buildDir)/output.%.lint: $(buildDir)/run-linter .FORCE
	@$(lintEnvVars) ./$< --output=$@ --lintBin=$(buildDir)/golangci-lint --packages='$*'
# end test and coverage artifacts

# start vendoring configuration
vendor-clean:
	rm -rf vendor/github.com/mongodb/grip/vendor/github.com/stretchr/testify/
	rm -rf vendor/github.com/mongodb/grip/vendor/github.com/pkg/errors/
	rm -rf vendor/github.com/evergreen-ci/gimlet/vendor/github.com/stretchr/testify/
	rm -rf vendor/github.com/evergreen-ci/gimlet/vendor/github.com/mongodb/grip/
	rm -rf vendor/github.com/evergreen-ci/gimlet/vendor/github.com/pkg/errors/
	find vendor/ -name "*.gif" -o -name "*.gz" -o -name "*.png" -o -name "*.ico" -o -name "*testdata*" | xargs rm -rf
phony += vendor-clean
# end vendoring configuration

# end cleanup targets
clean:
	rm -rf $(buildDir)
clean-results:
	rm -rf $(buildDir)/output.*
phony += clean clean-results
# end cleanup targets


# configure phony targets
.FORCE:
.PHONY: $(phony) .FORCE
