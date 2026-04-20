'use strict';

const { SurveyStatus, validateQuestions } = require('../models/survey');

class SurveyService {
  constructor(surveyRepository) {
    this.surveyRepository = surveyRepository;
  }

  async createSurvey({ title, description, questions }) {
    if (!title || title.trim().length === 0) {
      throw Object.assign(new Error('title is required'), { statusCode: 400 });
    }

    const validationError = validateQuestions(questions);
    if (validationError) {
      throw Object.assign(new Error(validationError), { statusCode: 400 });
    }

    return this.surveyRepository.createSurvey({ title: title.trim(), description, questions });
  }

  async getSurvey(id) {
    if (!id) {
      throw Object.assign(new Error('id is required'), { statusCode: 400 });
    }
    const survey = await this.surveyRepository.getSurvey(id);
    if (!survey) {
      throw Object.assign(new Error(`Survey ${id} not found`), { statusCode: 404 });
    }
    return survey;
  }

  async listSurveys({ status, limit, offset } = {}) {
    const parsedLimit = Math.min(parseInt(limit, 10) || 20, 100);
    const parsedOffset = parseInt(offset, 10) || 0;
    return this.surveyRepository.listSurveys({ status, limit: parsedLimit, offset: parsedOffset });
  }

  async activateSurvey(id) {
    const survey = await this.getSurvey(id);
    if (survey.status !== SurveyStatus.DRAFT) {
      throw Object.assign(
        new Error(`Cannot activate survey in status ${survey.status}. Survey must be in DRAFT status`),
        { statusCode: 422 }
      );
    }
    return this.surveyRepository.updateSurveyStatus(id, SurveyStatus.ACTIVE);
  }

  async closeSurvey(id) {
    const survey = await this.getSurvey(id);
    if (survey.status !== SurveyStatus.ACTIVE) {
      throw Object.assign(
        new Error(`Cannot close survey in status ${survey.status}. Survey must be in ACTIVE status`),
        { statusCode: 422 }
      );
    }
    return this.surveyRepository.updateSurveyStatus(id, SurveyStatus.CLOSED);
  }

  async deleteSurvey(id) {
    const survey = await this.getSurvey(id);
    if (survey.status !== SurveyStatus.DRAFT) {
      throw Object.assign(
        new Error(`Cannot delete survey in status ${survey.status}. Only DRAFT surveys can be deleted`),
        { statusCode: 422 }
      );
    }
    const deleted = await this.surveyRepository.deleteSurvey(id);
    if (!deleted) {
      throw Object.assign(new Error(`Survey ${id} not found`), { statusCode: 404 });
    }
    return deleted;
  }

  async submitResponse(surveyId, customerId, answers) {
    const survey = await this.getSurvey(surveyId);
    if (survey.status !== SurveyStatus.ACTIVE) {
      throw Object.assign(
        new Error(`Cannot submit response: survey is ${survey.status}, must be ACTIVE`),
        { statusCode: 422 }
      );
    }

    if (!answers || typeof answers !== 'object' || Object.keys(answers).length === 0) {
      throw Object.assign(new Error('answers must be a non-empty object'), { statusCode: 400 });
    }

    return this.surveyRepository.saveResponse({ surveyId, customerId, answers });
  }

  async getResponses(surveyId, limit) {
    const parsedLimit = Math.min(parseInt(limit, 10) || 50, 200);
    return this.surveyRepository.getResponses(surveyId, parsedLimit);
  }

  async getSurveyStats(surveyId) {
    await this.getSurvey(surveyId);
    const stats = await this.surveyRepository.getStats(surveyId);
    if (!stats) {
      throw Object.assign(new Error(`Stats for survey ${surveyId} not found`), { statusCode: 404 });
    }
    return stats;
  }
}

module.exports = SurveyService;
