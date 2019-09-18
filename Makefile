
OUT := build

Q = @

DEVNULL = /dev/null
ensure_dir=mkdir -p $(@D) 2> $(DEVNULL) || exit 0
CP=cp

.PHONY: all run

all: $(OUT)/komment

$(OUT)/komment: komment.go
ifdef Q
	@echo Building $@
endif
	$(Q)$(ensure_dir)
	$(Q)go build -o $@ $^
	$(Q)env GOOS=linux GOARCH=amd64 go build -o $@.linux_amd64 $^
