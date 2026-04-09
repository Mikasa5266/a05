import request from "../utils/request";

export const getQuestionBankMeta = () =>
  request({
    url: "/question-bank/meta",
    method: "get",
  });

export const getQuestionBankFilterOptions = (positionCode) =>
  request({
    url: "/question-bank/options",
    method: "get",
    params: {
      position_code: positionCode,
    },
  });

export const listQuestionBankQuestions = (params = {}) =>
  request({
    url: "/question-bank/questions",
    method: "get",
    params,
  });

export const getQuestionBankLists = (positionCode) =>
  request({
    url: "/question-bank/lists",
    method: "get",
    params: {
      position_code: positionCode,
    },
  });

export const drawQuestionBankQuestion = (params = {}) =>
  request({
    url: "/question-bank/random",
    method: "get",
    params,
  });

export const submitQuestionBankAnswer = (data) =>
  request({
    url: "/question-bank/answer",
    method: "post",
    data,
  });

export const getQuestionBankSolution = (questionId) =>
  request({
    url: `/question-bank/questions/${questionId}/solution`,
    method: "get",
  });

export const getQuestionBankPointSummary = (point, positionCode) =>
  request({
    url: `/question-bank/points/${encodeURIComponent(point)}/summary`,
    method: "get",
    params: {
      position_code: positionCode,
    },
  });

export const startQuestionBankAssessment = (data) =>
  request({
    url: "/question-bank/assessment/start",
    method: "post",
    data,
  });

export const submitQuestionBankAssessmentAnswer = (data) =>
  request({
    url: "/question-bank/assessment/answer",
    method: "post",
    data,
  });

export const completeQuestionBankAssessment = (assessmentId) =>
  request({
    url: `/question-bank/assessment/${assessmentId}/complete`,
    method: "post",
  });
