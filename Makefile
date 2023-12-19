NOW                     := $(shell date -u +'%Y-%m-%d_%TZ')
HEAD_SHA1               := $(shell git rev-parse HEAD)
HEAD_TAG                := $(shell git describe --tags | grep -e "^v" | sort | tail -1 | cut -b2-)
ifndef HEAD_TAG
HEAD_TAG := 0.0
$(warning No tags found in repository, using default version: $(HEAD_TAG)!)
endif

CODE_SIGN_CERT          := Manuel Koch Code Sign
APP_NAME                := Auto-WLAN

MACAPP_GO               := build/macapp.go
IMAGE_TO_GO             := ${GOPATH}/bin/2goarray

SPACE                   := $(subst ,, )
DARWIN_OS_VERSION       := $(subst $(SPACE),.,$(wordlist 1,2,$(subst ., ,$(shell sw_vers -productVersion))))
DARWIN_APP_ID           := com.manuel-koch.autowlan
DARWIN_APP_ICON_PNG     := assets/autowlan-black.png
DARWIN_DMG_BG_ICON_PNG  := assets/dmg-bg.png

DARWIN_ARM64_BINARY     := build/darwin-$(DARWIN_OS_VERSION)-arm64/autowlan.darwin-$(DARWIN_OS_VERSION)-arm64
DARWIN_ARM64_DIST_DIR   := dist/darwin-$(DARWIN_OS_VERSION)-arm64
DARWIN_ARM64_APP_BUNDLE := $(DARWIN_ARM64_DIST_DIR)/$(APP_NAME).app
DARWIN_ARM64_DMG        := $(DARWIN_ARM64_DIST_DIR)/$(APP_NAME)_v$(HEAD_TAG)_darwin_$(DARWIN_OS_VERSION)_arm64.dmg

DARWIN_AMD64_BINARY     := build/darwin-$(DARWIN_OS_VERSION)-amd64/autowlan.darwin-$(DARWIN_OS_VERSION)-amd64
DARWIN_AMD64_DIST_DIR   := dist/darwin-$(DARWIN_OS_VERSION)-amd64
DARWIN_AMD64_APP_BUNDLE := $(DARWIN_AMD64_DIST_DIR)/$(APP_NAME).app
DARWIN_AMD64_DMG        := $(DARWIN_AMD64_DIST_DIR)/$(APP_NAME)_v$(HEAD_TAG)_darwin_$(DARWIN_OS_VERSION)_amd64.dmg

# Choose between "debug" or "release" build type
BUILD_TYPE              := debug
DEBUG_BUILD_LDFLAGS     :=
RELEASE_BUILD_LDFLAGS   := -s -w
ifeq "$(BUILD_TYPE)" "debug"
BUILD_LDFLAGS := $(DEBUG_BUILD_LDFLAGS)
else
BUILD_LDFLAGS := $(RELEASE_BUILD_LDFLAGS)
endif

# Print a progress message with optional string ($1, $2)
define progress
@echo ""
@echo "********************************************************************"
$(if $1,@echo "*** $1",)
$(if $2,@echo "*** $2",)
endef

# Convert given string $1 from "snake_case" or "snake-case" to "UpperCamelCase"
define camelcase
$(shell echo "$1" | python -c "import sys; print(''.join(w.capitalize() for w in sys.stdin.read().replace('_','-').lstrip('-').split('-')))")
endef

$(MACAPP_GO):
	$(call progress,Fetching $@)
	curl -o $(MACAPP_GO) https://gist.githubusercontent.com/mholt/11008646c95d787c30806d3f24b2c844/raw/0c07883ba937f2d066d125ce3efd731adfd899d7/macapp.go

$(IMAGE_TO_GO):
	$(call progress,Building $@)
	go get github.com/cratonica/2goarray
	go install github.com/cratonica/2goarray

assets/%.go: $(IMAGE_TO_GO)
assets/%.go:
	$(call progress,Building $@,from     $<)
	echo "//+build linux darwin" > $@
	$(IMAGE_TO_GO) $(call camelcase,$(notdir $(basename $<))) assets < $< >> $@

assets/autowlan-black.go: assets/autowlan-black.png

assets/autowlan-black-disabled.go: assets/autowlan-black-disabled.png

autowlan.%: assets/autowlan-black.go assets/autowlan-black-disabled.go
	$(call progress,Building $(BUILD_TYPE) of $@)
	env GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=1 go build \
		-gcflags=all="-N -l" \
		-ldflags "$(BUILD_LDFLAGS) -X main.versionTag=$(HEAD_TAG) -X main.versionSha1=$(HEAD_SHA1) -X main.buildDate=$(NOW)" \
		-o $@ \
		.

%.app:
	$(call progress,Building $*,from $<)
	env GOOS=$(GOOS) GOARCH=$(GOARCH) go run $(MACAPP_GO) \
    		-assets $(dir $<) \
    		-bin $(notdir $<) \
			-icon $(DARWIN_APP_ICON_PNG) \
			-identifier $(DARWIN_APP_ID) \
			-name $(APP_NAME) \
			-o $(dir $@)
	plutil -replace CFBundleShortVersionString -string $(HEAD_TAG) $@/Contents/Info.plist
	plutil -replace LSUIElement -integer 1 $@/Contents/Info.plist

%.dmg:
	$(call progress,Building $*,from $<)
	if [[ -f $@ ]] ; then rm $@ ; fi
	create-dmg --volname $(notdir $@) --volicon $</Contents/Resources/icon.icns \
             --icon $(notdir $<) 110 150 \
             --app-drop-link 380 150 \
             --background $(DARWIN_DMG_BG_ICON_PNG) \
             $@ \
             $<
	$(call progress,Checksum $@)
	cd $(dir $@) && shasum -a 256 $(notdir $@) > $(notdir $@).sha256

.PHONY: %.signed
%.signed:
	$(call progress,Signing $*)
	security find-certificate -c "$(CODE_SIGN_CERT)" -p | openssl x509 -noout -text  -inform pem | grep -E "Validity|(Not (Before|After)\s*:)"
	codesign --verbose=4 --force --deep --sign "$(CODE_SIGN_CERT)" $*
	codesign --verbose=4 --display $*

$(DARWIN_AMD64_APP_BUNDLE): $(DARWIN_AMD64_BINARY) $(MACAPP_GO)
$(DARWIN_AMD64_APP_BUNDLE).signed: $(DARWIN_AMD64_APP_BUNDLE)
$(DARWIN_AMD64_DMG): $(DARWIN_AMD64_APP_BUNDLE) $(DARWIN_AMD64_APP_BUNDLE).signed

darwin_amd64_binary: GOOS=darwin
darwin_amd64_binary: GOARCH=amd64
darwin_amd64_binary: $(DARWIN_ARM64_BINARY)

darwin_amd64_bundle: GOOS=darwin
darwin_amd64_bundle: GOARCH=amd64
darwin_amd64_bundle: $(DARWIN_AMD64_APP_BUNDLE) $(DARWIN_AMD64_APP_BUNDLE).signed

darwin_amd64_dmg: GOOS=darwin
darwin_amd64_dmg: GOARCH=amd64
darwin_amd64_dmg: $(DARWIN_AMD64_DMG)

$(DARWIN_ARM64_APP_BUNDLE): $(DARWIN_ARM64_BINARY) $(MACAPP_GO)
$(DARWIN_ARM64_APP_BUNDLE).signed: $(DARWIN_ARM64_APP_BUNDLE)
$(DARWIN_ARM64_DMG): $(DARWIN_ARM64_APP_BUNDLE) $(DARWIN_ARM64_APP_BUNDLE).signed

darwin_arm64_binary: GOOS=darwin
darwin_arm64_binary: GOARCH=arm64
darwin_arm64_binary: $(DARWIN_ARM64_BINARY)

darwin_arm64_bundle: GOOS=darwin
darwin_arm64_bundle: GOARCH=arm64
darwin_arm64_bundle: $(DARWIN_ARM64_APP_BUNDLE) $(DARWIN_ARM64_APP_BUNDLE).signed

darwin_arm64_dmg: GOOS=darwin
darwin_arm64_dmg: GOARCH=arm64
darwin_arm64_dmg: $(DARWIN_ARM64_DMG)

.PHONY: darwin_all
darwin_all: darwin_amd64_dmg darwin_arm64_dmg

.PHONY: clean
clean::
	$(call progress,Cleaning)
	-rm -rf build/*
	-rm -rf dist/*
	-rm -f assets/*.go