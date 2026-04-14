import { computed, onBeforeUnmount, ref, unref, watch } from "vue";

export function useDeferredModelViewerMount(options = {}) {
  const {
    containerRef,
    enabled,
    startDelayMs = 360,
    idleTimeoutMs = 1800,
    rootMargin = "120px 0px",
    threshold = 0.15,
  } = options;

  const idleReady = ref(false);
  const visibleReady = ref(false);

  let idleHandle = null;
  let startDelayHandle = null;
  let observer = null;

  const clearStartDelay = () => {
    if (startDelayHandle !== null) {
      window.clearTimeout(startDelayHandle);
      startDelayHandle = null;
    }
  };

  const clearIdleTask = () => {
    if (idleHandle === null) return;

    if (typeof window !== "undefined" && "cancelIdleCallback" in window) {
      window.cancelIdleCallback(idleHandle);
    } else if (typeof window !== "undefined") {
      window.clearTimeout(idleHandle);
    }

    idleHandle = null;
  };

  const disconnectObserver = () => {
    if (!observer) return;
    observer.disconnect();
    observer = null;
  };

  const requestIdleMount = () => {
    if (typeof window === "undefined") {
      idleReady.value = true;
      return;
    }

    const markReady = () => {
      idleHandle = null;
      idleReady.value = true;
    };

    if ("requestIdleCallback" in window) {
      idleHandle = window.requestIdleCallback(markReady, {
        timeout: Math.max(0, Number(idleTimeoutMs) || 0),
      });
      return;
    }

    idleHandle = window.setTimeout(markReady, 0);
  };

  const setupVisibilityObserver = () => {
    if (typeof window === "undefined") {
      visibleReady.value = true;
      return;
    }

    const target = unref(containerRef);
    if (!target) return;

    if (!("IntersectionObserver" in window)) {
      visibleReady.value = true;
      return;
    }

    disconnectObserver();
    observer = new IntersectionObserver(
      (entries) => {
        const hit = entries.some((entry) => entry.isIntersecting);
        if (!hit) return;
        visibleReady.value = true;
        disconnectObserver();
      },
      {
        root: null,
        rootMargin,
        threshold,
      },
    );
    observer.observe(target);
  };

  const stopAll = () => {
    clearStartDelay();
    clearIdleTask();
    disconnectObserver();
  };

  const activate = () => {
    if (!unref(enabled)) {
      idleReady.value = false;
      visibleReady.value = false;
      stopAll();
      return;
    }

    idleReady.value = false;
    visibleReady.value = false;

    stopAll();
    setupVisibilityObserver();

    if (typeof window === "undefined") {
      idleReady.value = true;
      return;
    }

    startDelayHandle = window.setTimeout(
      () => {
        startDelayHandle = null;
        requestIdleMount();
      },
      Math.max(0, Number(startDelayMs) || 0),
    );
  };

  watch(
    () => unref(enabled),
    () => {
      activate();
    },
    { immediate: true },
  );

  watch(
    () => unref(containerRef),
    () => {
      if (!unref(enabled) || visibleReady.value) return;
      setupVisibilityObserver();
    },
  );

  onBeforeUnmount(() => {
    stopAll();
  });

  const shouldMountModelViewer = computed(() => {
    return Boolean(unref(enabled) && idleReady.value && visibleReady.value);
  });

  const modelViewerDeferred = computed(() => {
    return Boolean(unref(enabled) && !shouldMountModelViewer.value);
  });

  return {
    shouldMountModelViewer,
    modelViewerDeferred,
  };
}
