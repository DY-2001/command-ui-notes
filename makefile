build:
	@go build -o command-ui-notes.exe .

run: build
	@.\command-ui-notes.exe