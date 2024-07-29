import React, {useState, useMemo, useCallback, useEffect} from 'react'
import {Button, Modal, SegmentSection, Select, InlineFieldRow, SegmentInput, Drawer} from '@grafana/ui'
import {QueryEditorProps, SelectableValue} from '@grafana/data'
import {MacroType} from '@grafana/experimental'
import {FlightSQLDataSource} from '../datasource'
import {FlightSQLDataSourceOptions, SQLQuery, sqlLanguageDefinition, QUERY_FORMAT_OPTIONS} from '../types'
import {getSqlCompletionProvider, checkCasing} from './utils'

import {QueryEditorRaw} from './QueryEditorRaw'
import {BuilderView} from './BuilderView'
import {QueryHelp} from './QueryHelp'

export function QueryEditor(props: QueryEditorProps<FlightSQLDataSource, SQLQuery, FlightSQLDataSourceOptions>) {
  const {onChange, query, datasource} = props
  const [warningModal, showWarningModal] = useState(false)
  const [helpVisible, setHelpVisible] = useState(false)
  const [sqlInfo, setSqlInfo] = useState<any>()
  const [macros, setMacros] = useState<any>()
  const [rawEditor, setRawEditor] = useState<any>(false)
  const [format, setFormat] = useState<SelectableValue<string>>()
  const [fromRawSql, setFromSql] = useState(false)

  useEffect(() => {
    ;(async () => {
      const res = await datasource.getSQLInfo()
      const keywords = res?.frames[0].data.values[1][17] || []
      const numericFunctions = res?.frames[0].data.values[1][18] || []
      const stringFunctions = res?.frames[0].data.values[1][19] || []
      const systemFunctions = res?.frames[0].data.values[1][20] || []
      const sqlDateTimeFunctions = res?.frames[0].data.values[1][21] || []
      const functions = [...numericFunctions, ...stringFunctions, ...systemFunctions, ...sqlDateTimeFunctions]
      setSqlInfo({keywords: keywords, builtinFunctions: functions || []})
    })()
  }, [datasource])

  useEffect(() => {
    ;(async () => {
      const res = await datasource.getMacros()
      const prefix = `$__`
      const macroArr = res?.macros.map((m: any) => prefix.concat(m))
      const macros = macroArr.map((m: any) => ({text: m, name: m, id: m, type: MacroType.Value, args: []}));
      setMacros(macros)
    })()
  }, [datasource])

  const getTables = useCallback(async () => {
    const res = await datasource.getTables()
    return res?.frames[0].data.values[2].map((t: string) => ({
      name: checkCasing(t),
    }))
  }, [datasource])

  const getColumns = useCallback(
    async (table: any) => {
      let res
      if (table?.value) {
        res = await datasource.getColumns(table?.value)
      }
      return res?.frames[0].schema.fields
    },
    [datasource]
  )

  const completionProvider = useMemo(
    () =>
      getSqlCompletionProvider({
        getTables,
        getColumns,
        sqlInfo,
        macros,
      }),
    [getTables, getColumns, sqlInfo, macros]
  )

  useEffect(() => {
    // sets the format on the query on dropdown change
    if (format) {
      onChange({...query, format: format?.value})
    }

    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [format])

  useEffect(() => {
    // set the editor on the query
    onChange({...query, rawEditor: rawEditor})

    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [rawEditor])

  useEffect(() => {
    // get the format off the query on load
    if (query.format) {
      setFormat({value: query.format, label: query.format})
    }
    // set the default to table
    // if the user hadn't previously submitted a query with a format
    if (!query.format) {
      setFormat(QUERY_FORMAT_OPTIONS[1])
    }

    // check if a query has previously been sent from a
    // specific editor and default to that
    if (query.rawEditor) {
      setRawEditor(query.rawEditor)
    } else {
      setRawEditor(false)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const showHelp = () => {
    setHelpVisible(true)
  }
  const hideHelp = () => {
    setHelpVisible(false);
  }

  return (
    <>
      {warningModal && (
        <Modal
          title="Warning"
          closeOnBackdropClick={false}
          closeOnEscape={false}
          isOpen={warningModal}
          onDismiss={() => {
            showWarningModal(false)
          }}
        >
          {rawEditor
            ? 'By switching to the builder view you will not bring your current raw query over to the builder editor, you will have to fill it out again.'
            : 'By switching to the raw sql editor if you click to come back to the builder view you will need to refill your query.'}
          <Modal.ButtonRow>
            <Button fill="solid" size="md" variant="secondary" onClick={() => showWarningModal(!warningModal)}>
              Back
            </Button>
            <Button
              fill="solid"
              size="md"
              variant="destructive"
              onClick={() => {
                showWarningModal(!warningModal)
                setRawEditor(!rawEditor)
                setFromSql(rawEditor)
                if (rawEditor) {
                  query.queryText = ''
                }
              }}
            >
              Switch
            </Button>
          </Modal.ButtonRow>
        </Modal>
      )}
      {rawEditor ? (
        <QueryEditorRaw
          query={query}
          onChange={onChange}
          editorLanguageDefinition={{
            ...sqlLanguageDefinition,
            completionProvider,
          }}
        />
      ) : (
        <BuilderView query={props.query} datasource={datasource} onChange={onChange} fromRawSql={fromRawSql} />
      )}
      <div style={{width: '100%'}}>
        <InlineFieldRow style={{flexFlow: 'row', alignItems: 'center'}}>
          <SegmentSection label="Format As" fill={false}>
            <Select
              options={QUERY_FORMAT_OPTIONS}
              onChange={setFormat}
              value={query.format}
              width={15}
              placeholder="Table"
            />
          </SegmentSection>
          <Button style={{marginLeft: '5px'}} fill="outline" size="md" onClick={() => showWarningModal(!warningModal)}>
            {rawEditor ? 'Builder View' : 'Edit SQL'}
          </Button>
          <Button style={{marginLeft: '5px'}} fill="outline" size="md" onClick={showHelp}>
            Show Query Help
          </Button>
        </InlineFieldRow>
      </div>
      {!rawEditor && (
        <div style={{marginTop: '5px', whiteSpace: 'nowrap'}}>
          <SegmentSection label="Query Preview">
            <div style={{fontFamily: 'monospace', minWidth: '200px', fontSize: '12px'}}>
              <SegmentInput disabled value={query.queryText || ''} onChange={() => {}} />
            </div>
          </SegmentSection>{' '}
        </div>
      )}
      {helpVisible && <Drawer title="Query Help" onClose={hideHelp} closeOnMaskClick={true}><QueryHelp /></Drawer>}
    </>
  )
}
