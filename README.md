# Grafana Datasource Plugin

Datalayers datasource plugin for grafana. It is a plugin for Grafana that support Flight SQL APIs.

## Requirements

The plugin requires the user to run Grafana >=9.2.5.
The 10+ version is recommended.

## Local installation 

First of all. You need download the [plugin](https://github.com/datalayers-io/grafana-datalayers-datasource/releases).

And then create a directory eg custom/plugins/directory.

Now, you have two options to install on local:

1. Edit your `grafana.ini`

``` ini
[paths]
plugins = custom/plugins/directory/
```

2. Set the environment variable before Grafana started.

``` shell
GF_PATHS_PLUGINS=custom/plugins/directory/
```

## Install with Docker Run

```
docker run \
  -v $PWD/grafana-datalayers-datasource:/custom/plugins/directory/grafana-datalayers-datasource \
  -p 3344:3000 \
  --name grafana \
  grafana/grafana:latest
```

## Install with Docker-Compose
```
version: '3'
services:
  grafana:
    image: grafana/grafana:latest
    ports:
      - 3344:3000
    volumes: 
      - ./grafana-datalayers-datasource:/custom/plugins/directory/grafana-datalayers-datasource
```

## Usage

### Adding Datasource Plugin

1. Enter the `Configuration - Data Sources` page.
2. Click `Add new data source` button.
3. Search `Datalayers` keyword.
4. Click the Datalayers plugin to install.

### Configuring the Plugin

- **Host:** Provide the host:port of your Datalayers client.
- **Username/Password** Provide a username and password.
- **Require TLS/SSL:** Either enable or disable TLS based on the configuration of your client.
- **CA Cert** If you use yourself CA Cert file, Paste it in the textarea.
- **MetaData** Provide optional key, value pairs that you need sent to your Flight SQL client.


### Using the Query Builder

The default view is a query builder which is in active development:

- Begin by selecting the table from the dropdown.
- This will auto populate your available columns for your select statement. Use the **+** and **-** buttons to add or remove additional where statements.
- You can overwrite a dropdown field by typing in your desired value (e.g. `*`).
- The where field is a text entry where you can define any where clauses. Use the + and - buttons to add or remove additional where statements.
- You can switch to a raw SQL input by pressing the "Edit SQL" button. This will show you the query you have been building thus far and allow you to enter any query.
- Press the "Run query" button to see your results.
- From there you can add to dashboards and create any additional dashboards you like.

## Development

See [DEVELOPMENT.md](DEVELOPMENT.md).
