/**
 * Copyright (c) HashiCorp, Inc.
 * SPDX-License-Identifier: MPL-2.0
 */

export default {
  id: 'auth-form',
  initial: 'idle',
  on: {
    RESET: [
      {
        target: 'idle',
      },
    ],
    ERROR: [
      {
        target: 'error',
      },
    ],
  },
  states: {
    idle: {
      entry: ['clearError'],
      on: {
        SUBMIT: [
          {
            target: 'loading',
            cond: 'hasValue',
          },
          {
            target: 'error',
          },
        ],
      },
    },
    loading: {},
    error: {
      exit: ['clearError'],
      on: {
        TYPING: [
          {
            target: 'idle',
          },
        ],
        SUBMIT: [
          {
            target: 'loading',
            cond: 'hasValue',
          },
          {
            target: 'error',
          },
        ],
      },
    },
  },
};
