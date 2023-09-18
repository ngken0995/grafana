import { css, cx } from '@emotion/css';
import React, { useState, useEffect } from 'react';

import { GrafanaTheme2, StandardEditorProps } from '@grafana/data';
import { faro } from '@grafana/faro-web-sdk';
import { Button, Field, IconButton, useStyles2 } from '@grafana/ui';
import { FieldNamePicker } from '@grafana/ui/src/components/MatchersUI/FieldNamePicker';
import { LayerName } from 'app/core/components/Layers/LayerName';
import { ColorDimensionEditor, ScaleDimensionEditor } from 'app/features/dimensions/editors';

import { Options, ScatterSeriesConfig, defaultFieldConfig } from './panelcfg.gen';

export const ManualEditor = ({
  value,
  onChange,
  context,
}: StandardEditorProps<ScatterSeriesConfig[], any, Options>) => {
  const [selected, setSelected] = useState(0);
  const style = useStyles2(getStyles);

  const onFieldChange = (val: any | undefined, index: number, field: string) => {
    onChange(
      value.map((obj, i) => {
        if (i === index) {
          return { ...obj, [field]: val };
        }
        return obj;
      })
    );

    faro.api.pushEvent(
      'onFieldChange',
      { fieldVal: JSON.stringify(val), fieldIndex: index.toString(), fieldName: field },
      'xychart_panel'
    );
  };

  const createNewSeries = () => {
    onChange([
      ...value,
      {
        pointColor: {} as any,
        pointSize: defaultFieldConfig.pointSize,
      },
    ]);
    setSelected(value.length);

    faro.api.pushEvent('createNewSeries', { noOfSeries: (value.length + 1).toString() }, 'xychart_panel');
  };

  // Component-did-mount callback to check if a new series should be created
  useEffect(() => {
    if (!value?.length) {
      createNewSeries(); // adds a new series
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const onSeriesDelete = (index: number) => {
    onChange(value.filter((_, i) => i !== index));

    faro.api.pushEvent('onSeriesDelete', { noOfSeries: (value.length - 1).toString() }, 'xychart_panel');
  };

  // const { options } = context;

  const getRowStyle = (index: number) => {
    return index === selected ? `${style.row} ${style.sel}` : style.row;
  };

  return (
    <>
      <Button icon="plus" size="sm" variant="secondary" onClick={createNewSeries} className={style.marginBot}>
        Add series
      </Button>

      <div className={style.marginBot}>
        {value.map((series, index) => {
          return (
            // TODO: fix keyboard a11y
            // eslint-disable-next-line jsx-a11y/no-static-element-interactions
            <div key={`series/${index}`} className={getRowStyle(index)} onMouseDown={() => setSelected(index)}>
              <LayerName
                name={series.name ?? `Series ${index + 1}`}
                onChange={(v) => onFieldChange(v, index, 'name')}
              />

              <IconButton
                name="trash-alt"
                title={'remove'}
                className={cx(style.actionIcon)}
                onClick={() => onSeriesDelete(index)}
                tooltip="Delete series"
              />
            </div>
          );
        })}
      </div>

      {selected >= 0 && value[selected] && (
        <>
          <div key={`series/${selected}`}>
            <Field label={'X Field'}>
              <FieldNamePicker
                value={value[selected].x ?? ''}
                context={context}
                onChange={(field) => onFieldChange(field, selected, 'x')}
                item={{} as any}
              />
            </Field>
            <Field label={'Y Field'}>
              <FieldNamePicker
                value={value[selected].y ?? ''}
                context={context}
                onChange={(field) => onFieldChange(field, selected, 'y')}
                item={{} as any}
              />
            </Field>
            <Field label={'Point color'}>
              <ColorDimensionEditor
                value={value[selected].pointColor!}
                context={context}
                onChange={(field) => onFieldChange(field, selected, 'pointColor')}
                item={{} as any}
              />
            </Field>
            <Field label={'Point size'}>
              <ScaleDimensionEditor
                value={value[selected].pointSize!}
                context={context}
                onChange={(field) => onFieldChange(field, selected, 'pointSize')}
                item={{ settings: { min: 1, max: 100 } } as any}
              />
            </Field>
          </div>
        </>
      )}
    </>
  );
};

const getStyles = (theme: GrafanaTheme2) => ({
  marginBot: css`
    margin-bottom: 20px;
  `,
  row: css`
    padding: ${theme.spacing(0.5, 1)};
    border-radius: ${theme.shape.radius.default};
    background: ${theme.colors.background.secondary};
    min-height: ${theme.spacing(4)};
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 3px;
    cursor: pointer;

    border: 1px solid ${theme.components.input.borderColor};
    &:hover {
      border: 1px solid ${theme.components.input.borderHover};
    }
  `,
  sel: css`
    border: 1px solid ${theme.colors.primary.border};
    &:hover {
      border: 1px solid ${theme.colors.primary.border};
    }
  `,
  actionIcon: css`
    color: ${theme.colors.text.secondary};
    &:hover {
      color: ${theme.colors.text};
    }
  `,
});
