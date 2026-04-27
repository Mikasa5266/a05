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

  function buildQuery(params) {
    var entries = [];
    Object.keys(params || {}).forEach(function (key) {
      var value = params[key];
      if (value == null || value === "") return;
      entries.push(encodeURIComponent(key) + "=" + encodeURIComponent(String(value)));
    });
    return entries.length > 0 ? "?" + entries.join("&") : "";
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
      // ignore
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

    var moderationTypeEl = document.getElementById("moderation-type");
    var moderationKeywordEl = document.getElementById("moderation-keyword");
    var moderationRefreshBtn = document.getElementById("moderation-refresh-btn");
    var moderationMessageEl = document.getElementById("moderation-message");
    var moderationHeadEl = document.getElementById("moderation-head");
    var moderationBodyEl = document.getElementById("moderation-tbody");

    var auditKeywordEl = document.getElementById("audit-keyword");
    var auditRefreshBtn = document.getElementById("audit-refresh-btn");
    var auditMessageEl = document.getElementById("audit-message");
    var auditBodyEl = document.getElementById("audit-tbody");

    var patrolBtn = document.getElementById("patrol-btn");
    var patrolMessageEl = document.getElementById("patrol-message");
    var patrolLimitEl = document.getElementById("patrol-limit");

    if (logoutBtn) {
      logoutBtn.addEventListener("click", async function () {
        try {
          await apiFetch("/logout", { method: "POST" });
        } catch (error) {
          // ignore
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
      var protection = data.security_protection || {};
      var governance = data.illegal_content_governance || {};
      var contacts = data.security_contacts || {};
      var responsible = contacts.responsible_person || {};
      var emergency = contacts.emergency_contact || {};

      var total = Number(reportPipeline.total_reports || 0);
      var pending = Number(reportPipeline.pending_reports || 0);
      var processing = Number(reportPipeline.processing_reports || 0);
      var resolved = Number(reportPipeline.resolved_reports || 0);

      var kpiTotal = document.getElementById("kpi-total");
      var kpiPending = document.getElementById("kpi-pending");
      var kpiProcessing = document.getElementById("kpi-processing");
      var kpiResolved = document.getElementById("kpi-resolved");
      if (kpiTotal) kpiTotal.textContent = String(total);
      if (kpiPending) kpiPending.textContent = String(pending);
      if (kpiProcessing) kpiProcessing.textContent = String(processing);
      if (kpiResolved) kpiResolved.textContent = String(resolved);

      var realnameOk =
        !!realname.enabled &&
        !!realname.register_requires_phone &&
        !!realname.register_requires_id_no &&
        !!realname.register_requires_name;
      var retentionOk = !!retention.meets_six_months_min;
      var protectionOk = !!protection.sql_injection_guard && !!protection.xss_guard;
      var governanceOk =
        !!governance.keyword_filter && !!governance.patrol_enabled && !!governance.manual_disposal_enabled;

      var checkRealname = document.getElementById("check-realname");
      var checkRetention = document.getElementById("check-retention");
      var checkProtection = document.getElementById("check-protection");
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
      if (checkProtection) {
        checkProtection.innerHTML =
          "SQL 注入 / XSS 防护：<span class='" +
          (protectionOk ? "ok" : "danger") +
          "'>" +
          (protectionOk ? "已启用" : "未启用") +
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

    function reportActionButtons(report) {
      var statusLower = String(report.status || "").toLowerCase();
      if (statusLower !== "pending" && statusLower !== "processing") {
        return '<span class="muted">已结束</span>';
      }

      var buttons =
        '<div class="row">' +
        '<button class="button mini warn" data-action="processing" data-id="' +
        report.id +
        '">处理中</button>' +
        '<button class="button mini good" data-action="resolved" data-id="' +
        report.id +
        '">已处置</button>' +
        '<button class="button mini bad" data-action="rejected" data-id="' +
        report.id +
        '">驳回</button>';

      var targetType = String(report.target_type || "").toLowerCase();
      if (targetType === "post" || targetType === "comment" || targetType === "user") {
        var disposeAction = "delete_" + targetType;
        buttons +=
          '<button class="button mini danger-outline" data-action="' +
          disposeAction +
          '" data-id="' +
          report.id +
          '">删除对象并结案</button>';
      }
      buttons += "</div>";
      return buttons;
    }

    async function loadReports() {
      if (!tbody) return;
      tbody.innerHTML = "";

      var status = statusFilter ? statusFilter.value : "";
      setMessage(messageEl, "正在加载举报工单...", "muted");

      try {
        var query = buildQuery({ status: status });
        var data = await apiFetch("/reports" + query);
        var reports = Array.isArray(data.reports) ? data.reports : [];

        if (reports.length === 0) {
          setMessage(messageEl, "当前没有匹配的举报记录。", "muted");
          return;
        }

        var rows = reports.map(function (report) {
          var evidence = report.evidence || {};
          var evidenceText = Object.keys(evidence).length > 0 ? JSON.stringify(evidence) : "";
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
            reportActionButtons(report) +
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
        await Promise.all([loadCompliance(), loadReports(), loadAuditLogs()]);
      } catch (error) {
        setMessage(messageEl, "处理失败：" + error.message, "error");
      }
    }

    async function disposeReportTarget(id, action) {
      var note = window.prompt("请输入处置说明（会自动结案）", "");
      if (note === null) return;

      setMessage(messageEl, "正在删除目标内容并结案...", "muted");
      try {
        await apiFetch("/reports/" + encodeURIComponent(id) + "/dispose", {
          method: "POST",
          body: JSON.stringify({
            action: action,
            handle_note: note,
          }),
        });
        setMessage(messageEl, "已删除目标内容并完成结案。", "ok");
        await Promise.all([loadCompliance(), loadReports(), loadModerationList(), loadAuditLogs()]);
      } catch (error) {
        setMessage(messageEl, "处置失败：" + error.message, "error");
      }
    }

    function renderModerationRows(type, items) {
      if (!moderationHeadEl || !moderationBodyEl) return;

      if (type === "post") {
        moderationHeadEl.innerHTML =
          "<tr>" +
          "<th>ID</th><th>作者</th><th>标题</th><th>内容摘要</th><th>互动</th><th>时间</th><th>操作</th>" +
          "</tr>";
        moderationBodyEl.innerHTML = items
          .map(function (item) {
            return (
              "<tr>" +
              "<td>" + escapeHTML(item.id) + "</td>" +
              "<td>" + escapeHTML(item.author || "-") + " (uid:" + escapeHTML(item.user_id || "-") + ")</td>" +
              "<td>" + escapeHTML(item.title || "-") + "</td>" +
              "<td>" + escapeHTML(item.content || "-") + "</td>" +
              "<td>赞 " + escapeHTML(item.likes || 0) + " / 评 " + escapeHTML(item.comments || 0) + " / 阅 " + escapeHTML(item.views || 0) + "</td>" +
              "<td>" + escapeHTML(formatDate(item.created_at)) + "</td>" +
              '<td><button class="button mini danger-outline" data-type="post" data-id="' + escapeHTML(item.id) + '">删除帖子</button></td>' +
              "</tr>"
            );
          })
          .join("");
        return;
      }

      if (type === "comment") {
        moderationHeadEl.innerHTML =
          "<tr>" +
          "<th>ID</th><th>归属帖子</th><th>评论者</th><th>内容</th><th>时间</th><th>操作</th>" +
          "</tr>";
        moderationBodyEl.innerHTML = items
          .map(function (item) {
            return (
              "<tr>" +
              "<td>" + escapeHTML(item.id) + "</td>" +
              "<td>#"+ escapeHTML(item.post_id) + " " + escapeHTML(item.post_title || "-") + "</td>" +
              "<td>" + escapeHTML(item.author || "-") + " (uid:" + escapeHTML(item.user_id || "-") + ")</td>" +
              "<td>" + escapeHTML(item.content || "-") + "</td>" +
              "<td>" + escapeHTML(formatDate(item.created_at)) + "</td>" +
              '<td><button class="button mini danger-outline" data-type="comment" data-id="' + escapeHTML(item.id) + '">删除评论</button></td>' +
              "</tr>"
            );
          })
          .join("");
        return;
      }

      moderationHeadEl.innerHTML =
        "<tr>" +
        "<th>ID</th><th>用户名</th><th>角色</th><th>实名信息</th><th>邮箱/电话</th><th>时间</th><th>操作</th>" +
        "</tr>";
      moderationBodyEl.innerHTML = items
        .map(function (item) {
          return (
            "<tr>" +
            "<td>" + escapeHTML(item.id) + "</td>" +
            "<td>" + escapeHTML(item.username || "-") + "</td>" +
            "<td>" + escapeHTML(item.role || "-") + "</td>" +
            "<td>" + escapeHTML(item.real_name || "-") + " / " + (item.real_name_verified ? "已实名" : "未实名") + "</td>" +
            "<td>" + escapeHTML(item.email || "-") + "<br/><span class='muted'>" + escapeHTML(item.phone || "-") + "</span></td>" +
            "<td>" + escapeHTML(formatDate(item.created_at)) + "</td>" +
            '<td><button class="button mini danger-outline" data-type="user" data-id="' + escapeHTML(item.id) + '">删除用户</button></td>' +
            "</tr>"
          );
        })
        .join("");
    }

    async function loadModerationList() {
      if (!moderationTypeEl || !moderationHeadEl || !moderationBodyEl) return;

      var type = moderationTypeEl.value || "post";
      var keyword = moderationKeywordEl ? moderationKeywordEl.value.trim() : "";
      setMessage(moderationMessageEl, "正在加载内容列表...", "muted");

      var path = "/moderation/posts";
      var key = "posts";
      if (type === "comment") {
        path = "/moderation/comments";
        key = "comments";
      } else if (type === "user") {
        path = "/moderation/users";
        key = "users";
      }

      try {
        var data = await apiFetch(path + buildQuery({ keyword: keyword }));
        var items = Array.isArray(data[key]) ? data[key] : [];
        if (items.length === 0) {
          renderModerationRows(type, []);
          setMessage(moderationMessageEl, "当前无匹配数据。", "muted");
          return;
        }
        renderModerationRows(type, items);
        setMessage(moderationMessageEl, "内容列表加载成功。", "ok");
      } catch (error) {
        setMessage(moderationMessageEl, "加载失败：" + error.message, "error");
      }
    }

    async function deleteModerationItem(type, id) {
      var actionText = "删除";
      if (type === "post") actionText = "删除帖子";
      if (type === "comment") actionText = "删除评论";
      if (type === "user") actionText = "删除用户";
      var confirmed = window.confirm("确认" + actionText + " #" + id + "？此操作会真实落库。");
      if (!confirmed) return;

      setMessage(moderationMessageEl, "正在执行删除...", "muted");
      try {
        var path = "/moderation/" + type + "s/" + encodeURIComponent(id);
        await apiFetch(path, { method: "DELETE" });
        setMessage(moderationMessageEl, "删除成功。", "ok");
        await Promise.all([loadCompliance(), loadModerationList(), loadAuditLogs()]);
      } catch (error) {
        setMessage(moderationMessageEl, "删除失败：" + error.message, "error");
      }
    }

    async function loadAuditLogs() {
      if (!auditBodyEl) return;

      var keyword = auditKeywordEl ? auditKeywordEl.value.trim() : "";
      setMessage(auditMessageEl, "正在加载审计日志...", "muted");
      auditBodyEl.innerHTML = "";

      try {
        var data = await apiFetch("/audit-logs" + buildQuery({ keyword: keyword }));
        var logs = Array.isArray(data.audit_logs) ? data.audit_logs : [];
        if (logs.length === 0) {
          setMessage(auditMessageEl, "当前无匹配日志。", "muted");
          return;
        }

        auditBodyEl.innerHTML = logs
          .map(function (item) {
            var detailText = JSON.stringify(item.detail || {});
            if (detailText.length > 180) {
              detailText = detailText.slice(0, 180) + "...";
            }
            return (
              "<tr>" +
              "<td>" + escapeHTML(item.id) + "</td>" +
              "<td>" + escapeHTML(item.action || "-") + "</td>" +
              "<td>" + escapeHTML(item.outcome || "-") + "</td>" +
              "<td>" + escapeHTML((item.method || "-") + " " + (item.path || "-")) + "</td>" +
              "<td>" + escapeHTML(item.source_ip || "-") + "</td>" +
              "<td>" + escapeHTML(detailText) + "</td>" +
              "<td>" + escapeHTML(formatDate(item.created_at)) + "</td>" +
              "</tr>"
            );
          })
          .join("");

        setMessage(auditMessageEl, "审计日志加载成功。", "ok");
      } catch (error) {
        setMessage(auditMessageEl, "加载失败：" + error.message, "error");
      }
    }

    async function runPatrolNow() {
      if (!patrolBtn) return;
      var limit = patrolLimitEl ? Number(patrolLimitEl.value || 0) : 0;
      setMessage(patrolMessageEl, "正在执行巡查任务...", "muted");
      patrolBtn.disabled = true;
      try {
        var data = await apiFetch("/patrol/run", {
          method: "POST",
          body: JSON.stringify({
            limit: Number.isFinite(limit) && limit > 0 ? limit : 0,
          }),
        });
        var result = data.result || {};
        setMessage(
          patrolMessageEl,
          "巡查完成：扫描帖子 " +
            Number(result.scanned_posts || 0) +
            "，扫描评论 " +
            Number(result.scanned_comments || 0) +
            "，新建工单 " +
            Number(result.created_reports || 0),
          "ok"
        );
        await Promise.all([loadCompliance(), loadReports(), loadAuditLogs()]);
      } catch (error) {
        setMessage(patrolMessageEl, "巡查失败：" + error.message, "error");
      } finally {
        patrolBtn.disabled = false;
      }
    }

    if (refreshBtn) {
      refreshBtn.addEventListener("click", function () {
        Promise.all([loadCompliance(), loadReports(), loadModerationList(), loadAuditLogs()]);
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

        if (action === "processing" || action === "resolved" || action === "rejected") {
          handleReport(id, action);
          return;
        }
        if (action === "delete_post" || action === "delete_comment" || action === "delete_user") {
          disposeReportTarget(id, action);
        }
      });
    }

    if (moderationRefreshBtn) {
      moderationRefreshBtn.addEventListener("click", function () {
        loadModerationList();
      });
    }
    if (moderationTypeEl) {
      moderationTypeEl.addEventListener("change", function () {
        loadModerationList();
      });
    }
    if (moderationKeywordEl) {
      moderationKeywordEl.addEventListener("keydown", function (event) {
        if (event.key === "Enter") {
          loadModerationList();
        }
      });
    }
    if (moderationBodyEl) {
      moderationBodyEl.addEventListener("click", function (event) {
        var target = event.target;
        if (!target || target.tagName !== "BUTTON") return;
        var id = target.getAttribute("data-id");
        var type = target.getAttribute("data-type");
        if (!id || !type) return;
        deleteModerationItem(type, id);
      });
    }

    if (auditRefreshBtn) {
      auditRefreshBtn.addEventListener("click", function () {
        loadAuditLogs();
      });
    }
    if (auditKeywordEl) {
      auditKeywordEl.addEventListener("keydown", function (event) {
        if (event.key === "Enter") {
          loadAuditLogs();
        }
      });
    }

    if (patrolBtn) {
      patrolBtn.addEventListener("click", function () {
        runPatrolNow();
      });
    }

    await Promise.all([loadCompliance(), loadReports(), loadModerationList(), loadAuditLogs()]);
  }

  initLoginPage();
  initDashboardPage();
})();
