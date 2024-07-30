import React from 'react'
import {Card} from '@grafana/ui'

export const QueryHelp = () => (
  <div style={{padding: '10px 0'}}>
    <div>
      <h3>Format Options:</h3>
      <h4>Table(default):</h4>
      <p>Return any set of columns</p>
      <h3>Time series:</h3>
      <p>Return column named time (UTC in seconds or timestamp) return column(s) with numeric datatype as values</p>
      <p>Result sets of time series queries need to be sorted by time.</p>
    </div>
    <div>
      <h3>Supported Macros:</h3>
      <Card>
        <Card.Heading>$__interval</Card.Heading>
        <Card.Description>
        Get the appropriate time interval from Grafana panel then parse it like `1 seconds` <br />
        And the possible units are: milliseconds/seconds/minutes/hours/days <br />
        Example: <br />
        SELECT speed FROM test WHERE time &gt;= NOW() - INTERVAL &apos;$__interval&apos;<br />
        Be parsed as: <br />
        SELECT speed FROM test WHERE time &gt;= NOW() - INTERVAL &apos;1 seconds&apos;
        </Card.Description>
      </Card>

      <Card>
        <Card.Heading>$__dateBin</Card.Heading>
        <Card.Description>
        Group the data according to the appropriate time interval<br />
        Example: <br />
        SELECT $__dateBin(time) as timepoint FROM demo.test GROUP BY timepoint <br />
        Be parsed as: <br />
        SELECT date_bin(interval &apos;5 second&apos;, time, timestamp &apos;1970-01-01T00:00:00Z&apos;) as timepoint FROM demo.test GROUP BY timepoint
        </Card.Description>
      </Card>

      <Card>
        <Card.Heading>$__dateBinAlias</Card.Heading>
        <Card.Description>
        Group the data according to the appropriate time interval, use `_binned` suffix<br />
        Example: <br />
        SELECT $__dateBinAlias(time) FROM demo.test GROUP BY time_binned <br />
        Be parsed as: <br />
        SELECT date_bin(interval &apos;5 second&apos;, time, timestamp &apos;1970-01-01T00:00:00Z&apos;) as time_binned FROM demo.test GROUP BY ts_binned
        </Card.Description>
      </Card>

      <Card>
        <Card.Heading>$__timeFrom</Card.Heading>
        <Card.Description>
        Represent the start time selected for the Grafana panel<br />
        Example: <br />
        SELECT * FROM demo.test WHERE time &gt;= $__timeFrom<br />
        Be parsed as: <br />
        SELECT * FROM demo.test WHERE time &gt;= cast(&apos;2024-07-30T06:40:39Z&apos; as timestamp)
        </Card.Description>
      </Card>

      <Card>
        <Card.Heading>$__timeTo</Card.Heading>
        <Card.Description>
        Represent the end time selected for the Grafana panel<br />
        Example: <br />
        SELECT * FROM demo.test WHERE time &gt;= $__timeTo<br />
        Be parsed as: <br />
        SELECT * FROM demo.test WHERE time &gt;= cast(&apos;2024-07-30T06:40:39Z&apos; as timestamp)
        </Card.Description>
      </Card>

      <Card>
        <Card.Heading>$__timeFilter(time)</Card.Heading>
        <Card.Description>
        Represent the range time selected for the Grafana panel<br />
        Example: <br />
        SELECT * from demo.test WHERE $__timeFilter(time) <br />
        Be parsed as: <br />
        SELECT * from demo.test WHERE time &gt;= &apos;2024-07-30T07:36:07Z&apos; AND time &lt;= &apos;2024-07-30T10:36:07Z&apos; <br />
        Equal to: <br />
        SELECT * from demo.test WHERE time &gt;= $timeFrom AND time &lt;= $__timeTo
        </Card.Description>
      </Card>

      <Card>
        <Card.Heading>$__timeGroup(time, year)</Card.Heading>
        <Card.Description>
        Group the extracted partial time of the time field, the time field values: year/month/day/hour/minute<br />
        Example: <br />
        SELECT $__timeGroup(time, year) FROM demo.test<br />
        Be parsed as: <br />
        SELECT datepart(&apos;year&apos;, time),datepart(&apos;month&apos;, time),datepart(&apos;day&apos;, time),datepart(&apos;hour&apos;, time),datepart(&apos;minute&apos;, time) FROM demo.test
        </Card.Description>
      </Card>

      <Card>
        <Card.Heading>$__timeGroupAlias(time, year)</Card.Heading>
        <Card.Description>
        Group the extracted partial time of the time field, use alias prefix `time_`, the time field values: year/month/day/hour/minute<br />
        Example: <br />
        SELECT $__timeGroupAlias(time, year) FROM demo.test<br />
        Be parsed as: <br />
        SELECT datepart(&apos;year&apos;, time) as time_year,datepart(&apos;month&apos;, time) as time_month,datepart(&apos;day&apos;, time) as time_day,datepart(&apos;hour&apos;, time) as time_hour,datepart(&apos;minute&apos;, time) as time_minute FROM demo.test
        </Card.Description>
      </Card>

      <Card>
        <Card.Heading>$__timeRangeFrom(time)</Card.Heading>
        <Card.Description>
        Represent the start time selected for the Grafana panel<br />
        Example: <br />
        SELECT * FROM demo.test where $__timeRangeFrom(time)<br />
        Be parsed as: <br />
        SELECT * FROM demo.test where time &gt;= &apos;2024-07-30T07:29:46Z&apos;
        </Card.Description>
      </Card>

      <Card>
        <Card.Heading>$__timeRangeTo(time)</Card.Heading>
        <Card.Description>
        Represent the end time selected for the Grafana panel<br />
        Example: <br />
        SELECT * FROM demo.test where $__timeRangeTo(time)<br />
        Be parsed as: <br />
        SELECT * FROM demo.test where time &lt;= &apos;2024-07-30T07:29:46Z&apos;
        </Card.Description>
      </Card>

      

      <Card>
        <Card.Heading>$__timeRange(time)</Card.Heading>
        <Card.Description>
        Represent the range time selected for the Grafana panel <br />
        Example: <br />
        SELECT * FROM demo.test where $__timeRange(time)<br />
        Be parsed as: <br />
        SELECT * FROM demo.test where time &gt;= &apos;2024-07-30T07:37:41Z&apos; AND time &lt;= &apos;2024-07-30T10:37:41Z&apos; <br />
        Equal to: <br />
        SELECT * FROM demo.test where $__timeRangeFrom(time) AND $__timeRangeTo(time)
        </Card.Description>
      </Card>
    </div>
  </div>
)
