import React from 'react'
import {InlineSwitch, FieldSet, InlineField, SecretInput, Input, TextArea} from '@grafana/ui'
import {DataSourcePluginOptionsEditorProps} from '@grafana/data'
import {FlightSQLDataSourceOptions, SecureJsonData} from '../types'
import {
  onHostChange,
  onSecureChange,
  onUsernameChange,
  onPasswordChange,
  onResetPassword,
} from './utils'

export function ConfigEditor(props: DataSourcePluginOptionsEditorProps<FlightSQLDataSourceOptions, SecureJsonData>) {
  const {options, onOptionsChange} = props
  const {jsonData} = options
  const {secureJsonData, secureJsonFields} = options

  return (
    <div>
      <FieldSet label="Datalayers Connection" width={400}>
        <InlineField labelWidth={20} label="Host:Port">
          <Input
            width={40}
            name="host"
            type="text"
            value={jsonData.host || ''}
            placeholder="localhost:8360"
            onChange={(e) => onHostChange(e, options, onOptionsChange)}
          ></Input>
        </InlineField>

        <InlineField labelWidth={20} label="Username">
          <Input
            width={40}
            name="username"
            type="text"
            placeholder="username"
            onChange={(e) => onUsernameChange(e, options, onOptionsChange)}
            value={jsonData.username || ''}
          ></Input>
        </InlineField>
        <InlineField labelWidth={20} label="Password">
          <SecretInput
            width={40}
            name="password"
            type="text"
            value={secureJsonData?.password || ''}
            placeholder="****************"
            onChange={(e) => onPasswordChange(e, options, onOptionsChange)}
            onReset={() => onResetPassword(options, onOptionsChange)}
            isConfigured={secureJsonFields?.password}
          ></SecretInput>
        </InlineField>
        <InlineField labelWidth={20} label="Database">
          <Input
            width={40}
            name="database"
            type="text"
            placeholder="database name"
            onChange={(e) => {
              const jsonData = {
                ...options.jsonData,
                database: e.currentTarget.value,
              }
              onOptionsChange({ ...options, jsonData })
            }}
            value={jsonData.database || ''}
          ></Input>
        </InlineField>
        <InlineField labelWidth={20} label="Require TLS / SSL">
          <InlineSwitch
            label=""
            value={jsonData.secure}
            onChange={() => onSecureChange(options, onOptionsChange)}
            showLabel={false}
            disabled={false}
          />
        </InlineField>

        {
          jsonData.secure ? (
            <InlineField labelWidth={20} label="CA Cert">
              <TextArea value={secureJsonData?.tlsCACert} style={{ width: 400, minHeight: 120 }} placeholder="Begins with -----BEGIN CERTIFICATE-----" />
            </InlineField>
          ) : null
        }
        
      </FieldSet>
    </div>
  )
}
