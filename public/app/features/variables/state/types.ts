import { VariableType } from '@grafana/data';

import { VariableModel } from '../types';

export interface VariablesState extends Record<string, VariableModel> {}

export const initialVariablesState: VariablesState = {};

export interface VariableIdentifier {
  type: VariableType;
  id: string;
}

export interface KeyedVariableIdentifier {
  type: VariableType;
  id: string;
  rootStateKey: string;
}

export interface VariablePayload<T extends any = undefined> extends VariableIdentifier {
  data: T;
}

export interface AddVariable<T extends VariableModel = VariableModel> {
  global: boolean; // part of dashboard or global
  index: number; // the order in variables list
  model: T;
}
