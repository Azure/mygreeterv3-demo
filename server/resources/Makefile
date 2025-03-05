all: azuresdk alertrules
#This assumes ServiceResources.Parameters.json has already been templated.
MAKEFLAGS += -j2 # allows Makefile to run targets in parallel

.PHONY: alertrules
alertrules:
	./deployAzureResources.sh "alert-rules" false

.PHONY: clean
clean:
	@for dir in $$(find . -type d); do \
		(cd $$dir && if [ -f ".method_template_bicep.txt" ]; then \
			find . -name "*.bicep" -type f -delete; \
		fi); \
	done

.PHONY: azuresdk
azuresdk:
	./deployAzureResources.sh "azuresdk" true
