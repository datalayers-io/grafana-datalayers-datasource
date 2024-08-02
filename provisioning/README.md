# Introduction for grafana provisioning


## Config Datasource

1. First of all, build the plugin(frontend & backend).

2. Run `yarn server` for testing.

3. Visite the `http://localhost:3000`, login to home page.

4. Entry the `Home - Connections - Data sources` page, Click the `Datalayers` plugin.

5. Update the `Host:Port` field, replace the `your_host_ip` to real IP address, Click the `Save & test` button to testing connections. We will operate through port `8360` using the FlightSQL protocol to execute queries.



## Mock datas for Datalayers

We will operate through port `8361` using the HTTP protocol.

1. Create database `demo`

``` bash
curl -u"admin:public" -X POST \
http://127.0.0.1:8361/api/v1/sql \
-H 'Content-Type: application/binary' \
-d 'create database demo'
```

2. Create table `demo.sensor_info`

``` bash
curl -u"admin:public" -X POST \
http://127.0.0.1:8361/api/v1/sql?db=demo \
-H 'Content-Type: application/binary' \
-d 'CREATE TABLE sensor_info (
  ts TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  sn INT32 NOT NULL,
  speed int,
  longitude float,
  latitude float,
  timestamp KEY (ts)) PARTITION BY HASH(sn) PARTITIONS 2 ENGINE=TimeSeries;'
```

3. Write some data

``` bash
curl -u"admin:public" -X POST \
http://127.0.0.1:8361/api/v1/sql?db=demo \
-H 'Content-Type: application/binary' \
-d 'INSERT INTO sensor_info(sn, speed, longitude, latitude) VALUES(1, 120, 104.07, 30.59),(2, 120, 104.07, 30.59)'
```

## Query data by plugin

Enter the `Home - Explore` page and then start using the plugin to query the data.