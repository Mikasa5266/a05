<template>
  <div class="resume-type-modal" aria-hidden="true">
    <div class="resume-type-dialog">
      <header>
        <h3>选择文件类型</h3>
        <p>请选择你要上传的简历格式，再进入上传与解析流程。</p>
      </header>

      <div class="resume-type-grid">
        <button type="button" class="resume-type-card" data-type="pdf">
          <span class="resume-type-icon" aria-hidden="true">
            <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M8 3h6l4 4v13a1 1 0 0 1-1 1H8a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2Z" stroke="currentColor" stroke-width="1.8" />
              <path d="M14 3v4h4" stroke="currentColor" stroke-width="1.8" />
              <path d="M8.5 15.5h7M8.5 18h5" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" />
            </svg>
          </span>
          <span class="resume-type-label">PDF</span>
        </button>

        <button type="button" class="resume-type-card" data-type="word">
          <span class="resume-type-icon" aria-hidden="true">
            <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M8 3h6l4 4v13a1 1 0 0 1-1 1H8a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2Z" stroke="currentColor" stroke-width="1.8" />
              <path d="M14 3v4h4" stroke="currentColor" stroke-width="1.8" />
              <path d="m8.7 15 1.2 3 1.2-2 1.2 2 1.2-3" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" />
            </svg>
          </span>
          <span class="resume-type-label">Word</span>
        </button>

        <button type="button" class="resume-type-card" data-type="image">
          <span class="resume-type-icon" aria-hidden="true">
            <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <rect x="3" y="5" width="18" height="14" rx="2" stroke="currentColor" stroke-width="1.8" />
              <circle cx="9" cy="10" r="1.8" stroke="currentColor" stroke-width="1.8" />
              <path d="m6 17 4.2-4.2a1 1 0 0 1 1.4 0L16 17M13 14l1.5-1.5a1 1 0 0 1 1.4 0L19 15.6" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" />
            </svg>
          </span>
          <span class="resume-type-label">图片</span>
        </button>

        <button type="button" class="resume-type-card" data-type="scan">
          <span class="resume-type-icon" aria-hidden="true">
            <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M7 8V5a2 2 0 0 1 2-2h6a2 2 0 0 1 2 2v3" stroke="currentColor" stroke-width="1.8" />
              <rect x="4" y="8" width="16" height="7" rx="2" stroke="currentColor" stroke-width="1.8" />
              <path d="M8 14h8v7H8z" stroke="currentColor" stroke-width="1.8" />
              <path d="M9.5 17h5" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" />
            </svg>
          </span>
          <span class="resume-type-label">扫描件</span>
        </button>
      </div>

      <footer>
        <button type="button" class="secondary">取消</button>
      </footer>
    </div>
  </div>
</template>

<style scoped>
.resume-type-modal {
  position: fixed;
  inset: 0;
  background: rgba(12, 38, 58, 0.42);
  display: none;
  align-items: center;
  justify-content: center;
  padding: 20px;
  z-index: 30;
}

.resume-type-dialog {
  width: min(900px, calc(100vw - 120px));
  min-height: 620px;
  background: #ffffff;
  border: 1px solid #cfe0ee;
  border-radius: 14px;
  box-shadow: 0 10px 30px rgba(12, 44, 70, 0.2);
  padding: 24px;
  display: grid;
  gap: 20px;
  grid-template-rows: auto 1fr auto;
}

.resume-type-dialog header {
  text-align: center;
}

.resume-type-dialog h3 {
  margin: 0;
  font-size: 22px;
  color: #1b6f95;
}

.resume-type-dialog p {
  margin: 8px 0 0;
  font-size: 14px;
  color: #4a6178;
}

.resume-type-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  grid-template-rows: repeat(2, auto);
  gap: 16px;
  max-width: 620px;
  margin: 0 auto;
  width: 100%;
  align-content: center;
}

.resume-type-card {
  height: 150px;
  min-height: 150px;
  border-radius: 12px;
  border: 1px solid #cfe0ee;
  background: linear-gradient(180deg, #ffffff 0%, #f3f9ff 100%);
  color: #1b6f95;
  font-size: 16px;
  font-weight: 700;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.resume-type-card:hover {
  transform: translateY(-1px);
  border-color: #8bb7d2;
  box-shadow: 0 8px 16px rgba(27, 111, 149, 0.2);
}

.resume-type-icon {
  width: 34px;
  height: 34px;
  line-height: 0;
  color: #2a7eb1;
}

.resume-type-icon svg {
  width: 100%;
  height: 100%;
  display: block;
}

.resume-type-label {
  font-size: 16px;
  font-weight: 700;
  color: #1b6f95;
}

.resume-type-dialog footer {
  display: flex;
  justify-content: flex-end;
}

@media (max-width: 880px) {
  .resume-type-dialog {
    width: calc(100vw - 24px);
    min-height: auto;
    padding: 16px;
  }

  .resume-type-grid {
    max-width: 100%;
    gap: 12px;
  }

  .resume-type-card {
    height: 132px;
    min-height: 132px;
  }
}
</style>
