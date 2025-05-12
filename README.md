# DBQPDB Backend
## Installation
1. Install Go
2. Install Postgres (in any way that you could connect to)
3. Clone repository
4. Run the following in the root folder
```bash
go mod tidy
```
## Configuration
1. Rename `config\config.json.template` to `config\config.json` and change parameters according to your Postgres database
2. Run with `--setup` flag to create database structure
```bash
go run . --setup
```
3. Run normally onwards
```bash
go run .
```