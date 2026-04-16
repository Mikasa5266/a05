import {
  getHumanInvitations,
  getReceivedHumanInvitations,
  joinLiveInterview,
  startInterview,
} from "../api/interview";

export const TEST_START_THRESHOLD = 2;
export const TARGET_PARTICIPANTS = 4;
export const MAX_REMOTE_SLOTS = 3;

export function buildBackPath(currentPath) {
  const path = String(currentPath || "");
  if (path.startsWith("/enterprise/"))
    return "/enterprise/group-interview/workbench";
  if (path.startsWith("/university/"))
    return "/university/group-interview/workbench";
  return "/interview/group/workbench";
}

export function resolveInvitationId(route) {
  const fromParams = Number(route?.params?.id || 0);
  if (fromParams > 0) return fromParams;
  const fromQuery = Number(route?.query?.invitation_id || 0);
  return fromQuery > 0 ? fromQuery : 0;
}

export function resolveInvitationCode(route) {
  return String(route?.query?.invitation_code || "").trim();
}

export async function loadInvitationByID({
  invitationID,
  isStudent,
  fallbackInvitationCode,
}) {
  if (invitationID <= 0) {
    return {
      invitation: null,
      invitationCode: fallbackInvitationCode,
      interviewId: 0,
    };
  }

  let list = [];
  if (isStudent) {
    const [sentRes, receivedRes] = await Promise.all([
      getHumanInvitations(),
      getReceivedHumanInvitations(),
    ]);
    const sent = Array.isArray(sentRes?.invitations) ? sentRes.invitations : [];
    const received = Array.isArray(receivedRes?.invitations)
      ? receivedRes.invitations
      : [];
    list = [...sent, ...received];
  } else {
    const res = await getReceivedHumanInvitations();
    list = Array.isArray(res?.invitations) ? res.invitations : [];
  }

  const invitation =
    list.find((item) => Number(item.id) === invitationID) || null;
  return {
    invitation,
    invitationCode: String(
      invitation?.invitation_code || fallbackInvitationCode,
    ),
    interviewId: Number(invitation?.interview_id || 0),
  };
}

export async function ensureInterviewSession({
  invitation,
  isStudent,
  selfUserId,
  interviewId,
}) {
  if (!isStudent || !invitation || interviewId > 0) {
    return interviewId;
  }

  const currentUserID = Number(selfUserId || 0);
  const initiatorID = Number(
    invitation?.initiator_user_id || invitation?.student_id || 0,
  );
  if (currentUserID <= 0 || initiatorID <= 0 || currentUserID !== initiatorID) {
    return interviewId;
  }

  const payload = {
    position: invitation?.position || "群面模拟",
    difficulty: invitation?.difficulty || "campus_intern",
    mode: invitation?.mode || "comprehensive",
    style: invitation?.style || "gentle",
    company: invitation?.company || "",
    interview_mode: "human",
    invitation_id: Number(invitation?.id || 0),
  };

  const res = await startInterview(payload);
  const createdId = Number(res?.interview?.id || 0);
  return createdId > 0 ? createdId : interviewId;
}

export async function authorizeJoin({
  invitationID,
  invitationCode,
  fallbackInvitationCode,
}) {
  const res = await joinLiveInterview({
    invitation_id: Number(invitationID || 0),
    invitation_code: invitationCode || fallbackInvitationCode,
  });

  const session = res?.session || {};
  const roomId = String(session?.room_id || "").trim();
  if (!roomId) {
    throw new Error("房间授权信息无效");
  }

  return {
    roomId,
    invitationCode: String(
      session?.invitation_code || invitationCode || fallbackInvitationCode,
    ),
    interviewId: Number(session?.interview_id || 0),
  };
}

export function bindRemoteStreamToSlots({
  refs,
  peers,
  remoteMembers,
  selfUserId,
  maxRemoteSlots = MAX_REMOTE_SLOTS,
}) {
  const memberLookup = new Map(
    remoteMembers.map((item) => [item.userId, item]),
  );
  const orderedPeerEntries = Array.from(peers.entries())
    .filter(
      ([userId, state]) =>
        userId !== selfUserId &&
        state?.remoteStream &&
        memberLookup.has(userId),
    )
    .sort((a, b) => {
      const idxA = remoteMembers.findIndex((item) => item.userId === a[0]);
      const idxB = remoteMembers.findIndex((item) => item.userId === b[0]);
      return idxA - idxB;
    });

  for (let index = 0; index < maxRemoteSlots; index += 1) {
    const target = refs[index];
    if (!target) continue;
    const stream = orderedPeerEntries[index]?.[1]?.remoteStream || null;
    if (target.srcObject !== stream) {
      target.srcObject = stream;
      target.play?.().catch(() => {});
    }
  }
}

export function upsertRoomMember({ members, payload, selfUserId }) {
  const id = String(payload?.userId || payload?.user_id || "").trim();
  if (!id) return members;

  const displayName =
    String(
      payload?.senderName || payload?.sender_name || payload?.username || "",
    ).trim() || `成员 ${id}`;

  const nextMember = {
    userId: id,
    displayName,
    role: String(payload?.role || "")
      .trim()
      .toLowerCase(),
    isSelf: id === String(selfUserId || "").trim(),
  };

  const nextMembers = [...members];
  const idx = nextMembers.findIndex((item) => item.userId === id);
  if (idx >= 0) {
    nextMembers[idx] = { ...nextMembers[idx], ...nextMember };
  } else {
    nextMembers.push(nextMember);
  }
  return nextMembers;
}

export function removeRoomMember(members, userId) {
  const id = String(userId || "").trim();
  if (!id) return members;
  return members.filter((item) => item.userId !== id);
}

export function buildChatItem({ kind, text, fromSelf, senderName }) {
  const content = String(text || "").trim();
  if (!content) return null;
  return {
    id: `${Date.now()}-${Math.random().toString(16).slice(2)}`,
    kind,
    text: content,
    fromSelf,
    senderName: String(senderName || "系统").trim() || "系统",
    createdAt: new Date().toLocaleTimeString(),
  };
}
