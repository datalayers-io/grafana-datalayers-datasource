import { DataQueryResponse, MetricFindValue, DataSourceInstanceSettings, CoreApp, ScopedVars, VariableWithMultiSupport } from '@grafana/data'
import { frameToMetricFindValue, DataSourceWithBackend, getTemplateSrv } from '@grafana/runtime'
import { SQLQuery, FlightSQLDataSourceOptions, DEFAULT_QUERY } from './types'

import { lastValueFrom } from 'rxjs';

export class FlightSQLDataSource extends DataSourceWithBackend<SQLQuery, FlightSQLDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<FlightSQLDataSourceOptions>) {
    super(instanceSettings)
  }


async metricFindQuery(queryText: string, options?: any): Promise<MetricFindValue[]> {
  const target: SQLQuery = {
    refId: 'metricFindQuery',
    queryText,
    rawEditor: true,
    format: 'table'
  };
  return lastValueFrom(
    super.query({
      ...(options ?? {}), // includes 'range'
      targets: [target],
    })
  ).then(this.toMetricFindValue);
}

toMetricFindValue(rsp: DataQueryResponse): MetricFindValue[] {
  const data = rsp.data ?? [];
  // Create MetricFindValue object for all frames
  const values = data.map((d) => frameToMetricFindValue(d)).flat();
  // Filter out duplicate elements
  return values.filter((elm, idx, self) => idx === self.findIndex((t) => t.text === elm.text));

}

  getDefaultQuery(_: CoreApp): Partial<SQLQuery> {
    return DEFAULT_QUERY
  }

  quoteLiteral(value: string) {
    return "'" + value.replace(/'/g, "''") + "'"
  }

  interpolateVariable = (value: string | string[] | number, variable: VariableWithMultiSupport) => {
    if (typeof value === 'string') {
      if (variable?.multi || variable?.includeAll) {
        return this.quoteLiteral(value)
      } else {
        return String(value).replace(/'/g, "''")
      }
    }

    if (typeof value === 'number') {
      return value
    }

    if (Array.isArray(value)) {
      const quotedValues = value.map((v) => this.quoteLiteral(v))
      return quotedValues.join(',')
    }

    return value
  }

  replaceTimeUnit(timeText: String) {
    const replaced = timeText.replace(/(\d+)(ms|s|m|h|d)/g, (match, p1, p2) => {
      switch (p2) {
        case 'ms':
          return `${p1} milliseconds`;
        case 's':
          return `${p1} seconds`;
        case 'm':
          return `${p1} minutes`;
        case 'h':
          return `${p1} hours`;
        case 'd':
          return `${p1} days`;
        default:
          return match;
      }
    });
    return replaced;
  }

  overrideGrafanaVars(scopedVars: ScopedVars) {
    // Replace time unit to match Datalayers
    if (scopedVars.__interval.text) {
      const v = this.replaceTimeUnit(scopedVars.__interval.value);
      scopedVars.__interval.value = v;
    }
    
    return scopedVars;
  }

  applyTemplateVariables(query: SQLQuery, scopedVars: ScopedVars): Record<string, any> {
    this.overrideGrafanaVars(scopedVars);
    const interpolatedQuery: SQLQuery = {
      ...query,
      queryText: getTemplateSrv().replace(query.queryText, scopedVars, this.interpolateVariable),
    }
    return interpolatedQuery
  }

  getSQLInfo(): Promise<any> {
    return this.getResource('/flightsql/sql-info')
  }

  getTables(): Promise<any> {
    return this.getResource('/flightsql/tables')
  }

  getColumns(table: string): Promise<any> {
    return this.getResource(`/flightsql/columns?table=${table}`)
  }

  getMacros(): Promise<any> {
    return this.getResource('/plugin/macros')
  }
}
