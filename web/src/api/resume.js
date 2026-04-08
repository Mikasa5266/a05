import request from "../utils/request";

export function parseResume(formData, source = "web") {
  const data = formData instanceof FormData ? formData : new FormData();
  if (!(formData instanceof FormData) && formData?.file) {
    data.append("file", formData.file);
  }
  if (source) {
    data.append("source", source);
  }

  return request({
    url: "/resume/parse",
    method: "post",
    data,
    headers: {
      "Content-Type": "multipart/form-data",
    },
  });
}

export function getLatestResumeAnalysis() {
  return request({
    url: "/resume/latest",
    method: "get",
  });
}
