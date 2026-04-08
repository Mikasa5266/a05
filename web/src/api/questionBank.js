import request from "../utils/request";

export function getQuestionBankPositions() {
  return request({
    url: "/question-bank/positions",
    method: "get",
  });
}

export function getPositionQuestionList(positionCode, params = {}) {
  return request({
    url: `/question-bank/positions/${encodeURIComponent(positionCode)}/questions`,
    method: "get",
    params,
  });
}

export function generateResumeQuestionList(resumeResultID, payload = {}) {
  return request({
    url: `/question-bank/resume/${encodeURIComponent(resumeResultID)}/questions`,
    method: "post",
    params: {
      difficulty: payload?.difficulty || "",
      limit: payload?.limit || 12,
    },
  });
}

export function getQuestionById(questionID) {
  return request({
    url: `/question-bank/questions/${encodeURIComponent(questionID)}`,
    method: "get",
  });
}

export function evaluateQuestion(questionID, answer) {
  return request({
    url: `/question-bank/questions/${encodeURIComponent(questionID)}/evaluate`,
    method: "post",
    data: { answer },
  });
}

export function setQuestionFavorite(questionID, isFavorite) {
  return request({
    url: `/question-bank/questions/${encodeURIComponent(questionID)}/favorite`,
    method: "post",
    data: { is_favorite: Boolean(isFavorite) },
  });
}

export function listFavoriteQuestions(params = {}) {
  return request({
    url: "/question-bank/favorites",
    method: "get",
    params,
  });
}

export function markQuestionWrong(questionID, note = "") {
  return request({
    url: `/question-bank/questions/${encodeURIComponent(questionID)}/wrong`,
    method: "post",
    data: { note },
  });
}

export function clearQuestionWrong(questionID) {
  return request({
    url: `/question-bank/questions/${encodeURIComponent(questionID)}/wrong`,
    method: "delete",
  });
}

export function listWrongQuestions(params = {}) {
  return request({
    url: "/question-bank/wrong-questions",
    method: "get",
    params,
  });
}
