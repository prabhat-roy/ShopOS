'use strict';

class SurveyController {
  constructor(surveyService) {
    this.surveyService = surveyService;
    this.createSurvey = this.createSurvey.bind(this);
    this.getSurvey = this.getSurvey.bind(this);
    this.listSurveys = this.listSurveys.bind(this);
    this.activateSurvey = this.activateSurvey.bind(this);
    this.closeSurvey = this.closeSurvey.bind(this);
    this.deleteSurvey = this.deleteSurvey.bind(this);
    this.submitResponse = this.submitResponse.bind(this);
    this.getResponses = this.getResponses.bind(this);
    this.getSurveyStats = this.getSurveyStats.bind(this);
  }

  async createSurvey(req, res) {
    const { title, description, questions } = req.body;

    if (!title) {
      return res.status(400).json({ error: 'Bad Request', message: 'title is required' });
    }

    try {
      const survey = await this.surveyService.createSurvey({ title, description, questions });
      return res.status(201).json(survey);
    } catch (err) {
      return this._handleError(res, err);
    }
  }

  async getSurvey(req, res) {
    const { id } = req.params;
    try {
      const survey = await this.surveyService.getSurvey(id);
      return res.status(200).json(survey);
    } catch (err) {
      return this._handleError(res, err);
    }
  }

  async listSurveys(req, res) {
    const { status, limit, offset } = req.query;
    try {
      const result = await this.surveyService.listSurveys({ status, limit, offset });
      return res.status(200).json(result);
    } catch (err) {
      return this._handleError(res, err);
    }
  }

  async activateSurvey(req, res) {
    const { id } = req.params;
    try {
      await this.surveyService.activateSurvey(id);
      return res.status(204).send();
    } catch (err) {
      return this._handleError(res, err);
    }
  }

  async closeSurvey(req, res) {
    const { id } = req.params;
    try {
      await this.surveyService.closeSurvey(id);
      return res.status(204).send();
    } catch (err) {
      return this._handleError(res, err);
    }
  }

  async deleteSurvey(req, res) {
    const { id } = req.params;
    try {
      await this.surveyService.deleteSurvey(id);
      return res.status(204).send();
    } catch (err) {
      return this._handleError(res, err);
    }
  }

  async submitResponse(req, res) {
    const { id: surveyId } = req.params;
    const { customerId, answers } = req.body;

    if (!answers) {
      return res.status(400).json({ error: 'Bad Request', message: 'answers is required' });
    }

    try {
      const response = await this.surveyService.submitResponse(surveyId, customerId, answers);
      return res.status(201).json(response);
    } catch (err) {
      return this._handleError(res, err);
    }
  }

  async getResponses(req, res) {
    const { id: surveyId } = req.params;
    const { limit } = req.query;
    try {
      // Verify survey exists (throws 404 if not)
      await this.surveyService.getSurvey(surveyId);
      const responses = await this.surveyService.getResponses(surveyId, limit);
      return res.status(200).json({ surveyId, responses });
    } catch (err) {
      return this._handleError(res, err);
    }
  }

  async getSurveyStats(req, res) {
    const { id: surveyId } = req.params;
    try {
      const stats = await this.surveyService.getSurveyStats(surveyId);
      return res.status(200).json({ surveyId, ...stats });
    } catch (err) {
      return this._handleError(res, err);
    }
  }

  _handleError(res, err) {
    const statusCode = err.statusCode || 500;
    const message = err.message || 'Internal server error';

    if (statusCode === 500) {
      console.error('[SurveyController] Internal error:', err);
    }

    return res.status(statusCode).json({
      error: statusCode === 400 ? 'Bad Request'
        : statusCode === 404 ? 'Not Found'
        : statusCode === 422 ? 'Unprocessable Entity'
        : 'Internal Server Error',
      message,
    });
  }
}

module.exports = SurveyController;
