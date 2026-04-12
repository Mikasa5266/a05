import { defineAsyncComponent } from "vue";

const Login = defineAsyncComponent(() => import("../../views/Login.vue"));
const Forbidden = defineAsyncComponent(
  () => import("../../views/Forbidden.vue"),
);

const publicMeta = {
  requiresAuth: false,
  roles: ["public"],
};

export const commonRoutes = [
  {
    path: "/",
    redirect: "/student/login",
    meta: publicMeta,
  },
  {
    path: "/student/login",
    name: "StudentLogin",
    component: Login,
    meta: publicMeta,
  },
  {
    path: "/enterprise/login",
    name: "EnterpriseLogin",
    component: Login,
    meta: publicMeta,
  },
  {
    path: "/university/login",
    name: "UniversityLogin",
    component: Login,
    meta: publicMeta,
  },
  {
    path: "/403",
    name: "Forbidden",
    component: Forbidden,
    meta: publicMeta,
  },
  {
    path: "/login",
    redirect: "/student/login",
    meta: publicMeta,
  },
  {
    path: "/dashboard",
    redirect: "/student/dashboard",
    meta: publicMeta,
  },
];
