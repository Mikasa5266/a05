(function () {
  "use strict";

  var API_BASE = "/admin-platform/api";
  var TOKEN_KEY = "admin_platform_token";

  function getToken() {
    return window.localStorage.getItem(TOKEN_KEY) || "";
  }

  function setToken(token) {
    if (!token) {
      window.localStorage.removeItem(TOKEN_KEY);
      return;
    }
    window.localStorage.setItem(TOKEN_KEY, token);
  }

  async function apiFetch(path, options, withAuth) {
    var opts = options || {};
    var headers = opts.headers || {};
    headers["Content-Type"] = "application/json";
    if (withAuth !== false) {
      var token = getToken();
      if (token) {
        headers["X-Admin-Token"] = token;
      }
    }

    var response = await fetch(API_BASE + path, {
      method: opts.method || "GET",
      headers: headers,
      body: opts.body,
      credentials: "same-origin",
    });

    var data = {};
    try {
      data = await response.json();
    } catch (e) {
      data = {};
    }

    if (!response.ok) {
      var msg = data.error || "request failed";
      var error = new Error(msg);
      error.status = response.status;
      throw error;
    }
    return data;
  }

  function formatDate(value) {
    if (!value) return "-";
    var dt = new Date(value);
    if (Number.isNaN(dt.getTime())) return "-";
    return dt.toLocaleString("zh-CN");
  }

  function escapeHTML(input) {
    var str = String(input == null ? "" : input);
    return str
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;")
      .replace(/'/g, "&#039;");
  }

  function statusTag(status) {
    var s = String(status || "").toLowerCase();
    if (s !== "pending" && s !== "processing" && s !== "resolved" && s !== "rejected") {
      s = "pending";
    }
    return '<span class="tag ' + s + '">' + s + "</span>";
  }

  function setMessage(el, text, type) {
    if (!el) return;
    el.textContent = text || "";
    el.className = "msg";
    if (type === "ok") el.classList.add("ok");
    if (type === "error") el.classList.add("error");
    if (type === "muted") el.classList.add("muted");
  }

  async function ensureLogin() {
    var token = getToken();
    if (!token) {
      window.location.href = "/admin-platform/login";
      return null;
    }
    try {
      return await apiFetch("/me");
    } catch (error) {
      setToken("");
      window.location.href = "/admin-platform/login";
      return null;
    }
  }

  async function initLoginPage() {
    var form = document.getElementById("login-form");
    if (!form) return;

    var msg = document.getElementById("login-message");
    var btn = document.getElementById("login-btn");

    try {
      var me = await apiFetch("/me");
      if (me && me.username) {
        window.location.href = "/admin-platform/dashboard";
        return;
      }
    } catch (error) {
      // not logged in
    }

    form.addEventListener("submit", async function (event) {
      event.preventDefault();
      if (btn) btn.disabled = true;

      var username = document.getElementById("username").value.trim();
      var password = document.getElementById("password").value;

      try {
        var result = await apiFetch(
          "/login",
          {
            method: "POST",
            body: JSON.stringify({
              username: username,
              password: password,
            }),
          },
          false
        );
        setToken(result.token || "");
        setMessage(msg, "登录成功，正在进入控制台...", "ok");
        window.location.href = "/admin-platform/dashboard";
      } catch (error) {
        setMessage(msg, "登录失败：" + error.message, "error");
      } finally {
        if (btn) btn.disabled = false;
      }
    });
  }

  async function initDashboardPage() {
    var tbody = document.getElementById("reports-tbody");
    if (!tbody) return;

    var me = await ensureLogin();
    if (!me) return;

    var usernameEl = document.getElementById("admin-username");
    if (usernameEl) {
      usernameEl.textContent = "管理员：" + (me.username || "-");
    }

    var messageEl = document.getElementById("dashboard-message");
    var statusFilter = document.getElementById("status-filter");
    var refreshBtn = document.getElementById("refresh-btn");
    var logoutBtn = document.getElementById("logout-btn");

    if (logoutBtn) {
      logoutBtn.addEventListener("click", async function () {
        try {
          await apiFetch("/logout", { method: "POST" });
        } catch (error) {
          // ignore logout failure
        } finally {
          setToken("");
          window.location.href = "/admin-platform/login";
        }
      });
    }

    async function loadCompliance() {
      var data = await apiFetch("/compliance/overview");
      var reportPipeline = data.report_pipeline || {};
      var realname = data.realname_verification || {};
      var retention = data.log_retention || {};
      var governance = data.illegal_content_governance || {};
      var contacts = data.security_contacts || {};
      var responsible = contacts.responsible_person || {};
      var emergency = contacts.emergency_contact || {};

      var total = Number(reportPipeline.total_reports || 0);
      var pending = Number(reportPipeline.pending_reports || 0);
      var processing = Number(reportPipeline.processing_reports || 0);

      var kpiTotal = document.getElementById("kpi-total");
      var kpiPending = document.getElementById("kpi-pending");
      var kpiProcessing = document.getElementById("kpi-processing");
      if (kpiTotal) kpiTotal.textContent = String(total);
      if (kpiPending) kpiPending.textContent = String(pending);
      if (kpiProcessing) kpiProcessing.textContent = String(processing);

      var realnameOk =
        !!realname.enabled &&
        !!realname.register_requires_phone &&
        !!realname.register_requires_id_no &&
        !!realname.register_requires_name;
      var retentionOk = !!retention.meets_six_months_min;
      var governanceOk =
        !!governance.keyword_filter && !!governance.patrol_enabled && !!governance.manual_disposal_enabled;

      var checkRealname = document.getElementById("check-realname");
      var checkRetention = document.getElementById("check-retention");
      var checkGovernance = document.getElementById("check-governance");
      var checkContact = document.getElementById("check-contact");
      if (checkRealname) {
        checkRealname.innerHTML =
          "实名校验（手机号/身份证）：<span class='" +
          (realnameOk ? "ok" : "danger") +
          "'>" +
          (realnameOk ? "已启用" : "未启用") +
          "</span>";
      }
      if (checkRetention) {
        checkRetention.innerHTML =
          "访问/操作日志留存：<span class='" +
          (retentionOk ? "ok" : "danger") +
          "'>" +
          (retentionOk ? "满足至少 6 个月（当前 " + Number(retention.days || 0) + " 天）" : "不满足") +
          "</span>";
      }
      if (checkGovernance) {
        checkGovernance.innerHTML =
          "违法信息过滤、巡查和处置：<span class='" +
          (governanceOk ? "ok" : "danger") +
          "'>" +
          (governanceOk
            ? "已启用（巡查间隔 " + Number(governance.patrol_interval_minutes || 0) + " 分钟）"
            : "未启用") +
          "</span>";
      }
      if (checkContact) {
        var hasContacts =
          String(responsible.name || "").trim() !== "" && String(emergency.name || "").trim() !== "";
        checkContact.innerHTML =
          "安全责任人与应急联系人：<span class='" +
          (hasContacts ? "ok" : "danger") +
          "'>" +
          (hasContacts ? "已配置" : "未配置") +
          "</span>";
      }

      var complaintEl = document.getElementById("contact-complaint");
      var responsibleEl = document.getElementById("contact-responsible");
      var emergencyEl = document.getElementById("contact-emergency");
      if (complaintEl) {
        complaintEl.textContent =
          "举报投诉电话：" + (contacts.complaint_phone || "-") + "；举报邮箱：" + (contacts.complaint_email || "-");
      }
      if (responsibleEl) {
        responsibleEl.textContent =
          "安全责任人：" +
          (responsible.name || "-") +
          "，电话：" +
          (responsible.phone || "-") +
          "，邮箱：" +
          (responsible.email || "-");
      }
      if (emergencyEl) {
        emergencyEl.textContent =
          "应急联系人：" +
          (emergency.name || "-") +
          "，电话：" +
          (emergency.phone || "-") +
          "，邮箱：" +
          (emergency.email || "-");
      }
    }

    async function loadReports() {
      if (!tbody) return;
      tbody.innerHTML = "";

      var status = statusFilter ? statusFilter.value : "";
      setMessage(messageEl, "正在加载举报工单...", "muted");

      try {
        var query = status ? "?status=" + encodeURIComponent(status) : "";
        var data = await apiFetch("/reports" + query);
        var reports = Array.isArray(data.reports) ? data.reports : [];

        if (reports.length === 0) {
          setMessage(messageEl, "当前没有匹配的举报记录。", "muted");
          return;
        }

        var rows = reports.map(function (report) {
          var statusLower = String(report.status || "").toLowerCase();
          var evidence = report.evidence || {};
          var evidenceText = Object.keys(evidence).length > 0 ? JSON.stringify(evidence) : "";
          var actionButtons = "";
          if (statusLower === "pending" || statusLower === "processing") {
            actionButtons =
              '<div class="row">' +
              '<button class="button warn" data-action="processing" data-id="' +
              report.id +
              '">处理中</button>' +
              '<button class="button good" data-action="resolved" data-id="' +
              report.id +
              '">已处置</button>' +
              '<button class="button bad" data-action="rejected" data-id="' +
              report.id +
              '">驳回</button>' +
              "</div>";
          } else {
            actionButtons = '<span class="muted">已结束</span>';
          }

          return (
            "<tr>" +
            "<td>" +
            escapeHTML(report.id) +
            "</td>" +
            "<td>" +
            statusTag(report.status) +
            "</td>" +
            "<td>" +
            escapeHTML(String(report.target_type || "") + "#" + String(report.target_id || 0)) +
            "</td>" +
            "<td>" +
            escapeHTML(report.reason || "-") +
            "</td>" +
            "<td>" +
            escapeHTML(report.description || "-") +
            (evidenceText ? "<br/><span class='muted'>证据：" + escapeHTML(evidenceText) + "</span>" : "") +
            (report.handle_note
              ? "<br/><span class='muted'>处理备注：" + escapeHTML(report.handle_note) + "</span>"
              : "") +
            "</td>" +
            "<td>" +
            escapeHTML(report.contact || "-") +
            "</td>" +
            "<td>" +
            escapeHTML(formatDate(report.created_at)) +
            "</td>" +
            "<td>" +
            actionButtons +
            "</td>" +
            "</tr>"
          );
        });

        tbody.innerHTML = rows.join("");
        setMessage(messageEl, "举报工单加载成功。", "ok");
      } catch (error) {
        setMessage(messageEl, "加载失败：" + error.message, "error");
      }
    }

    async function handleReport(id, status) {
      var note = window.prompt("请输入处理备注（可留空）", "");
      if (note === null) return;

      setMessage(messageEl, "正在提交处理结果...", "muted");
      try {
        await apiFetch("/reports/" + encodeURIComponent(id) + "/handle", {
          method: "POST",
          body: JSON.stringify({
            status: status,
            handle_note: note,
          }),
        });
        setMessage(messageEl, "处理成功。", "ok");
        await Promise.all([loadCompliance(), loadReports()]);
      } catch (error) {
        setMessage(messageEl, "处理失败：" + error.message, "error");
      }
    }

    if (refreshBtn) {
      refreshBtn.addEventListener("click", function () {
        Promise.all([loadCompliance(), loadReports()]);
      });
    }

    if (statusFilter) {
      statusFilter.addEventListener("change", function () {
        loadReports();
      });
    }

    if (tbody) {
      tbody.addEventListener("click", function (event) {
        var target = event.target;
        if (!target || target.tagName !== "BUTTON") return;
        var id = target.getAttribute("data-id");
        var action = target.getAttribute("data-action");
        if (!id || !action) return;
        handleReport(id, action);
      });
    }

    await Promise.all([loadCompliance(), loadReports()]);
  }

  initLoginPage();
  initDashboardPage();
})();
