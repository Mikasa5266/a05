import request from "../utils/request";

export const getPracticeMeta = () =>
  request({
    url: "/practice/meta",
    method: "get",
  });

export const getPracticeFilterOptions = (positionCode) =>
  request({
    url: "/practice/bank/options",
    method: "get",
    params: {
      position_code: positionCode,
    },
  });

export const listPracticeQuestions = (params = {}) =>
  request({
    url: "/practice/questions",
    method: "get",
    params,
  });

export const getPracticeQuestionLists = (positionCode) =>
  request({
    url: "/practice/question-lists",
    method: "get",
    params: {
      position_code: positionCode,
    },
  });

export const getPracticeQuestionsOfList = (listId, params = {}) =>
  request({
    url: `/practice/question-lists/${listId}/questions`,
    method: "get",
    params,
  });

export const getPracticeSpecialties = (positionCode) =>
  request({
    url: "/practice/specialties",
    method: "get",
    params: {
      position_code: positionCode,
    },
  });

export const drawPracticeQuestion = (params = {}) =>
  request({
    url: "/practice/question/random",
    method: "get",
    params,
  });

export const getPracticeQuestion = (questionId) =>
  request({
    url: `/practice/question/${questionId}`,
    method: "get",
  });

export const submitPracticeAnswer = (data) =>
  request({
    url: "/practice/answer",
    method: "post",
    data,
  });

export const getPracticeSolution = (questionId) =>
  request({
    url: `/practice/question/${questionId}/solution`,
    method: "get",
  });

export const getPracticePointSummary = (positionCode, point) =>
  request({
    url: "/practice/point/summary",
    method: "get",
    params: {
      position_code: positionCode,
      point,
    },
  });

export const startPracticeAssessment = (data) =>
  request({
    url: "/practice/assessment/start",
    method: "post",
    data,
  });

export const submitPracticeAssessmentAnswer = (data) =>
  request({
    url: "/practice/assessment/answer",
    method: "post",
    data,
  });

export const completePracticeAssessment = (assessmentId) =>
  request({
    url: `/practice/assessment/${assessmentId}/complete`,
    method: "post",
  });

export const getPracticeIntegrationSnapshot = (positionCode) =>
  request({
    url: "/practice/integration/snapshot",
    method: "get",
    params: {
      position_code: positionCode,
    },
    headers: {
      "X-Skip-Error-Toast": "true",
    },
  });

export const submitPracticeIntegrationFeedback = (data) =>
  request({
    url: "/practice/integration/feedback",
    method: "post",
    data,
  });

export const getPracticeWrongs = (params = {}) =>
  request({
    url: "/practice/wrongs",
    method: "get",
    params,
  });

export const getPracticeWrongRemedial = (wrongId) =>
  request({
    url: `/practice/wrongs/${wrongId}/remedial`,
    method: "get",
  });

export const deletePracticeWrong = (wrongId) =>
  request({
    url: `/practice/wrongs/${wrongId}`,
    method: "delete",
  });

export const togglePracticeWrongFavorite = (wrongId) =>
  request({
    url: `/practice/wrongs/${wrongId}/favorite`,
    method: "post",
  });

export const togglePracticeQuestionFavorite = (questionId) =>
  request({
    url: `/practice/favorites/${questionId}`,
    method: "post",
  });

export const getPracticeDashboard = (positionCode) =>
  request({
    url: "/practice/dashboard",
    method: "get",
    params: {
      position_code: positionCode,
    },
  });

export const importPracticeQuestionBank = (data) =>
  request({
    url: "/practice/questionbank/import",
    method: "post",
    data,
  });

export const exportPracticeRecords = () =>
  request({
    url: "/practice/records/export",
    method: "get",
    responseType: "blob",
  });
