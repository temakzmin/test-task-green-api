(function () {
  // Production: "/api/v1" behind reverse proxy.
  // Local without reverse proxy: use "http://localhost:5050/api/v1".
  const API_BASE = "/api/v1";

  const responseEl = document.getElementById("response");
  const statusBadge = document.getElementById("statusBadge");

  const controls = {
    idInstance: document.getElementById("idInstance"),
    apiTokenInstance: document.getElementById("apiTokenInstance"),
    toggleApiTokenVisibility: document.getElementById("toggleApiTokenVisibility"),
    smChatId: document.getElementById("smChatId"),
    smMessage: document.getElementById("smMessage"),
    sfChatId: document.getElementById("sfChatId"),
    sfUrl: document.getElementById("sfUrl"),
    btnGetSettings: document.getElementById("btnGetSettings"),
    btnGetStateInstance: document.getElementById("btnGetStateInstance"),
    btnSendMessage: document.getElementById("btnSendMessage"),
    btnSendFileByUrl: document.getElementById("btnSendFileByUrl")
  };

  function setStatus(state, text) {
    statusBadge.classList.remove("success", "error");
    if (state === "success") statusBadge.classList.add("success");
    if (state === "error") statusBadge.classList.add("error");
    statusBadge.textContent = text;
  }

  function setResponse(payload) {
    responseEl.value = typeof payload === "string" ? payload : JSON.stringify(payload, null, 2);
  }

  function credentials() {
    return {
      idInstance: controls.idInstance.value.trim(),
      apiTokenInstance: controls.apiTokenInstance.value.trim()
    };
  }

  function assertCredentials() {
    const creds = credentials();
    if (!creds.idInstance || !creds.apiTokenInstance) {
      throw new Error("Заполните idInstance и ApiTokenInstance");
    }
    return creds;
  }

  function setButtonsDisabled(disabled) {
    controls.btnGetSettings.disabled = disabled;
    controls.btnGetStateInstance.disabled = disabled;
    controls.btnSendMessage.disabled = disabled;
    controls.btnSendFileByUrl.disabled = disabled;
  }

  function setTokenVisibility(isVisible) {
    controls.apiTokenInstance.type = isVisible ? "text" : "password";
    controls.toggleApiTokenVisibility.classList.toggle("is-visible", isVisible);
    controls.toggleApiTokenVisibility.setAttribute("aria-pressed", String(isVisible));
    controls.toggleApiTokenVisibility.setAttribute("aria-label", isVisible ? "Скрыть токен" : "Показать токен");
    controls.toggleApiTokenVisibility.title = isVisible ? "Скрыть токен" : "Показать токен";
  }

  async function callApi(path, payload) {
    setButtonsDisabled(true);
    setStatus("idle", "loading");

    try {
      const res = await fetch(`${API_BASE}${path}`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json"
        },
        body: JSON.stringify(payload)
      });

      const contentType = res.headers.get("content-type") || "";
      const rawBody = await res.text();
      let body = rawBody;

      if (contentType.includes("application/json") && rawBody.trim() !== "") {
        try {
          body = JSON.parse(rawBody);
        } catch {
          body = rawBody;
        }
      }

      setResponse(body);

      if (res.ok) {
        setStatus("success", `success ${res.status}`);
      } else {
        setStatus("error", `error ${res.status}`);
      }
    } catch (error) {
      setStatus("error", "network error");
      setResponse({
        error: {
          message: error instanceof Error ? error.message : String(error)
        }
      });
    } finally {
      setButtonsDisabled(false);
    }
  }

  controls.btnGetSettings.addEventListener("click", async function () {
    try {
      await callApi("/settings", assertCredentials());
    } catch (error) {
      setStatus("error", "validation error");
      setResponse({ error: { message: error.message } });
    }
  });

  controls.btnGetStateInstance.addEventListener("click", async function () {
    try {
      await callApi("/state", assertCredentials());
    } catch (error) {
      setStatus("error", "validation error");
      setResponse({ error: { message: error.message } });
    }
  });

  controls.btnSendMessage.addEventListener("click", async function () {
    try {
      const creds = assertCredentials();
      const chatId = controls.smChatId.value.trim();
      const message = controls.smMessage.value.trim();

      if (!chatId || !message) {
        throw new Error("Для sendMessage заполните chatId и message");
      }

      await callApi("/send-message", {
        ...creds,
        chatId,
        message
      });
    } catch (error) {
      setStatus("error", "validation error");
      setResponse({ error: { message: error.message } });
    }
  });

  controls.btnSendFileByUrl.addEventListener("click", async function () {
    try {
      const creds = assertCredentials();
      const chatId = controls.sfChatId.value.trim();
      const urlFile = controls.sfUrl.value.trim();

      if (!chatId || !urlFile) {
        throw new Error("Для sendFileByUrl заполните chatId и urlFile");
      }

      await callApi("/send-file-by-url", {
        ...creds,
        chatId,
        urlFile
      });
    } catch (error) {
      setStatus("error", "validation error");
      setResponse({ error: { message: error.message } });
    }
  });

  controls.toggleApiTokenVisibility.addEventListener("click", function () {
    setTokenVisibility(controls.apiTokenInstance.type === "password");
  });

  setTokenVisibility(false);

  setResponse({
    info: "Введите параметры подключения и вызовите один из методов"
  });
})();
