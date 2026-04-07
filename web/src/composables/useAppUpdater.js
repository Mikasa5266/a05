import { computed, onBeforeUnmount, ref } from "vue";
import native from "../utils/native";

export function useAppUpdater() {
  const status = ref({
    stage: "idle",
    message: "",
    version: "",
    percent: 0,
    transferred: 0,
    total: 0,
    bytesPerSecond: 0,
  });

  const isUpdating = computed(() => status.value.stage === "downloading");
  const isReadyToInstall = computed(() => status.value.stage === "downloaded");
  const progressPercent = computed(() => Number(status.value.percent || 0));

  const stopListening = native.onUpdaterStatus((payload) => {
    status.value = {
      ...status.value,
      ...payload,
    };
  });

  onBeforeUnmount(() => {
    stopListening();
  });

  const checkForUpdates = () => native.checkForUpdates();
  const quitAndInstallUpdate = () => native.quitAndInstallUpdate();

  return {
    status,
    isUpdating,
    isReadyToInstall,
    progressPercent,
    checkForUpdates,
    quitAndInstallUpdate,
  };
}
