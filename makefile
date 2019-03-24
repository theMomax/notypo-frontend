all: help

# from https://gist.github.com/prwhite/8168133
help: ## This help dialog.
	@IFS=$$'\n' ; \
	help_lines=(`fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//'`); \
	for help_line in $${help_lines[@]}; do \
		IFS=$$'#' ; \
		help_split=($$help_line) ; \
		help_command=`echo $${help_split[0]} | sed -e 's/^ *//' -e 's/ *$$//'` ; \
		help_info=`echo $${help_split[2]} | sed -e 's/^ *//' -e 's/ *$$//'` ; \
		printf "%-30s %s\n" $$help_command $$help_info ; \
	done

init: fonts css ## Initializes or updates the following resources:
##				- fonts
##				- css

fonts: ## Initializes or updates all fonts
	wget -O style/fonts/zillaslab.ttf https://github.com/google/fonts/raw/master/ofl/zillaslab/ZillaSlab-Regular.ttf
	
css: ## Initializes or updates all css-files
	rm -f style/css/*.css
	for i in ./style/css/*; \
	do \
		lessc $$i $${i%.less}.css; \
	done

test: ## Runs all package-tests.
	go test ./wasm/comparison/...

run: ## Starts a webserver for development. This command requires github.com/dennwc/dom/cmd/wasm-server.
	wasm-server -apps wasm -main notypo