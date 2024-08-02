import {DataSourcePluginOptionsEditorProps, PluginType} from '@grafana/data'

import {FlightSQLDataSource} from './datasource'
import {FlightSQLDataSourceOptions, SQLQuery} from './types'

export const mockDatasource = new FlightSQLDataSource({
  id: 1,
  uid: 'datalayers-flightsql-id',
  type: 'datalayers-flightsql-datasource',
  name: 'FlightSQL Data Source',
  readOnly: false,
  jsonData: {},
  access: 'proxy',
  meta: {
    id: 'datalayersio-datasource',
    module: '',
    name: 'FlightSQL Data Source',
    type: PluginType.datasource,
    alerting: true,
    backend: true,
    baseUrl: 'public/plugins/datalayersio-datasource',
    info: {
      description: '',
      screenshots: [],
      updated: '',
      version: '',
      logos: {
        small: '',
        large: '',
      },
      author: {
        name: '',
      },
      links: [],
    },
  },
})

export const mockDatasourceOptions: DataSourcePluginOptionsEditorProps<FlightSQLDataSourceOptions> = {
  options: {
    id: 1,
    uid: '1',
    orgId: 1,
    name: 'Timestream',
    typeLogoUrl: '',
    type: '',
    access: '',
    url: '',
    user: '',
    basicAuth: false,
    basicAuthUser: '',
    database: '',
    isDefault: false,
    jsonData: {
      host: '',
      secure: true,
      username: '',
      selectedAuthType: '',
      metadata: [],
    },
    secureJsonFields: {
      token: false,
      password: false,
    },
    readOnly: false,
    withCredentials: false,
    typeName: '',
  },
  onOptionsChange: jest.fn(),
}

export const mockQuery: SQLQuery = {
  queryText: 'show databases',
  refId: '',
  format: 'table',
  rawEditor: true,
  table: '',
  columns: [],
  wheres: [],
  orderBy: '',
  groupBy: '',
  limit: '',
}
