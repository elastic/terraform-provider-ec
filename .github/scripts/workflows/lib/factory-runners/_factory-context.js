'use strict';

const FACTORY_CONSTANTS = {
  'change-factory': require('../intake/change-factory-constants.js'),
  'code-factory': require('../intake/code-factory-constants.js'),
  'research-factory': require('../intake/research-factory-constants.js'),
  'reproducer-factory': require('../intake/reproducer-factory-constants.js'),
};

function getFactoryName() {
  const factory = process.env.FACTORY_NAME;
  if (!factory) {
    throw new Error('FACTORY_NAME environment variable is required.');
  }
  if (!Object.prototype.hasOwnProperty.call(FACTORY_CONSTANTS, factory)) {
    throw new Error(`Unknown FACTORY_NAME: ${factory}. Expected one of: ${Object.keys(FACTORY_CONSTANTS).join(', ')}.`);
  }
  return factory;
}

function getFactoryConstants() {
  return FACTORY_CONSTANTS[getFactoryName()];
}

function getFactoryModule() {
  // Loaded on demand so the context-dir helpers do not depend on factory-issue-shared.js,
  // which lands with the gate runners.
  const { createFactoryIssueModule } = require('../factory-issue-shared.js');
  const c = getFactoryConstants();
  return createFactoryIssueModule({
    branchPrefix: c.ISSUE_BRANCH_PREFIX,
    factoryLabel: c.FACTORY_LABEL,
    issueOpenedNotEligibleReason: c.ISSUE_OPENED_NOT_ELIGIBLE_REASON,
    duplicateLinkageMode: c.DUPLICATE_LINKAGE_MODE,
  });
}

function getFactoryContextDir() {
  return `/tmp/${getFactoryName()}-context`;
}

module.exports = { getFactoryName, getFactoryConstants, getFactoryModule, getFactoryContextDir };
