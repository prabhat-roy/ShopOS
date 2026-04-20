'use strict';

const QuestionType = Object.freeze({
  TEXT: 'TEXT',
  SINGLE_CHOICE: 'SINGLE_CHOICE',
  MULTIPLE_CHOICE: 'MULTIPLE_CHOICE',
  RATING: 'RATING',
  NPS: 'NPS',
});

const SurveyStatus = Object.freeze({
  DRAFT: 'DRAFT',
  ACTIVE: 'ACTIVE',
  CLOSED: 'CLOSED',
});

const ALL_QUESTION_TYPES = Object.values(QuestionType);
const ALL_SURVEY_STATUSES = Object.values(SurveyStatus);

function isValidQuestionType(type) {
  return ALL_QUESTION_TYPES.includes(type);
}

function isValidSurveyStatus(status) {
  return ALL_SURVEY_STATUSES.includes(status);
}

/**
 * Validate a questions array.
 * Each question must have: id, type, text
 * SINGLE_CHOICE and MULTIPLE_CHOICE must have options array.
 */
function validateQuestions(questions) {
  if (!Array.isArray(questions) || questions.length === 0) {
    return 'questions must be a non-empty array';
  }
  for (const q of questions) {
    if (!q.id) return 'each question must have an id';
    if (!q.text) return 'each question must have text';
    if (!isValidQuestionType(q.type)) {
      return `invalid question type: ${q.type}. Valid: ${ALL_QUESTION_TYPES.join(', ')}`;
    }
    if (
      (q.type === QuestionType.SINGLE_CHOICE || q.type === QuestionType.MULTIPLE_CHOICE) &&
      (!Array.isArray(q.options) || q.options.length === 0)
    ) {
      return `question ${q.id} of type ${q.type} must have options array`;
    }
  }
  return null;
}

module.exports = {
  QuestionType,
  SurveyStatus,
  ALL_QUESTION_TYPES,
  ALL_SURVEY_STATUSES,
  isValidQuestionType,
  isValidSurveyStatus,
  validateQuestions,
};
