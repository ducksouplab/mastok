(() => {
  // front/src/js/lib/fingerprint.js
  var VERSION = "1";
  var LOCAL_KEY = "mastok_random_id";
  var getWebGLInfo = () => {
    let output = "";
    let gl = document.createElement("canvas").getContext("webgl");
    if (gl) {
      output += `${gl.getParameter(gl.VERSION)}${gl.getParameter(gl.VENDOR)}${gl.getParameter(gl.RENDERER)}`;
      debug = gl.getExtension("WEBGL_debug_renderer_info");
      if (debug) {
        output += gl.getParameter(debug.UNMASKED_RENDERER_WEBGL);
      }
    }
    return output;
  };
  var getLocalId = () => {
    let localId = localStorage.getItem(LOCAL_KEY);
    if (localId) {
      return localId;
    }
    localId = crypto.randomUUID();
    localStorage.setItem(LOCAL_KEY, localId);
    return localId;
  };
  var concatenateMetrics = () => {
    let metrics = VERSION;
    metrics += getLocalId();
    metrics += `${serverHash}`;
    metrics += navigator.userAgent;
    metrics += navigator.deviceMemory;
    metrics += navigator.hardwareConcurrency;
    metrics += navigator.language;
    metrics += navigator.maxTouchPoints;
    metrics += navigator.pdfViewerEnabled;
    metrics += window.screen.width;
    metrics += window.screen.height;
    metrics += window.screen.pixelDepth;
    metrics += devicePixelRatio;
    metrics += (/* @__PURE__ */ new Date()).getTimezoneOffset();
    if (Intl) {
      metrics += Intl.DateTimeFormat().resolvedOptions().timeZone;
    }
    metrics += getWebGLInfo();
    return metrics;
  };
  var digest = async (message) => {
    const msgUint8 = new TextEncoder().encode(message);
    const hashBuffer = await crypto.subtle.digest("SHA-512", msgUint8);
    const hashArray = Array.from(new Uint8Array(hashBuffer));
    const hashHex = hashArray.map((b) => b.toString(16).padStart(2, "0")).join("");
    return hashHex;
  };
  var fingerprint = async () => {
    const metrics = concatenateMetrics();
    hash = await digest(metrics);
    return hash;
  };
  var fingerprint_default = fingerprint;

  // front/src/js/join.js
  var state = {
    starting: false
  };
  var looseJSONParse = (str) => {
    try {
      return JSON.parse(str);
    } catch (error) {
      console.error(error);
    }
  };
  var containers = [
    "consent-container",
    "grouping-container",
    "waiting-container",
    "joining-container",
    "instructions-container",
    "pending-container",
    "full-container",
    "unavailable-container",
    "landing-failed-container"
  ];
  var show = (id) => {
    const target = document.getElementById(id);
    if (target) {
      target.style.display = "";
    }
  };
  var hide = (id) => {
    const target = document.getElementById(id);
    if (target) {
      target.style.display = "none";
    }
  };
  var showOnly = (id) => {
    for (const c of containers) {
      c === id ? show(c) : hide(c);
    }
  };
  var processConsent = (html) => {
    let output = html.replaceAll("<a ", '<a target="_blank"');
    output = html.replaceAll(
      '<input type="checkbox" disabled',
      '<input type="checkbox"'
    );
    return output;
  };
  var submitConsent = (ws) => {
    const checkboxes = document.querySelectorAll('#consent-container input[type="checkbox"]');
    let accepted = true;
    for (const c of checkboxes) {
      accepted = accepted && c.checked;
    }
    if (accepted) {
      hide("alert-container");
      ws.send(JSON.stringify({ kind: "Agree" }));
    } else {
      show("alert-container");
    }
  };
  var start = function(slug) {
    const wsProtocol = window.location.protocol === "https:" ? "wss" : "ws";
    const pathPrefixhMatch = /(.*)join/.exec(window.location.pathname);
    const pathPrefix = pathPrefixhMatch[1];
    const wsUrl = `${wsProtocol}://${window.location.host}${pathPrefix}ws/join?slug=${slug}`;
    const ws = new WebSocket(wsUrl);
    ws.onopen = async () => {
      const uid = await fingerprint_default();
      ws.send(JSON.stringify({ kind: "Land", payload: uid }));
    };
    ws.onclose = (event) => {
      if (!state.starting) {
        showOnly("unavailable-container");
      }
      console.log(event);
    };
    ws.onerror = (event) => {
      showOnly("unavailable-container");
      console.log(event);
    };
    ws.onmessage = (event) => {
      const { kind, payload } = looseJSONParse(event.data);
      console.log(kind, payload);
      if (kind === "Consent") {
        document.querySelector("#consent-container").innerHTML = processConsent(payload);
        hide("alert-container");
        const lis = document.querySelectorAll("#consent-container li");
        for (const li of lis) {
          li.addEventListener("click", () => {
            const checkbox = li.querySelector("input");
            if (!checkbox.disabled) {
              value = checkbox.checked;
              checkbox.checked = !value;
            }
          });
        }
        const submitButton = document.querySelector("#consent-container button");
        if (submitButton) {
          submitButton.addEventListener("click", () => submitConsent(ws));
        }
        showOnly("consent-container");
      } else if (kind === "Grouping") {
        const [question, ...answers] = payload.split("\n");
        const action = answers.pop();
        document.getElementById("grouping-question").innerText = question;
        document.getElementById("grouping-submit").innerText = action;
        const groupingAnswers = document.getElementById("grouping-answers");
        for (let a of answers) {
          const [text, _] = a.split(":");
          const answerEl = document.createElement("div");
          answerEl.classList.add("form-check");
          answerEl.innerHTML = `<input class="form-check-input" type="radio" value="${text}" name="group" id="answer-${text}" required>
        <label class="form-check-label" for="answer-${text}">${text}</label>`;
          groupingAnswers.append(answerEl);
        }
        const formChecks = document.querySelectorAll("#grouping-answers input");
        for (const c of formChecks) {
          c.addEventListener("change", () => {
            document.getElementById("grouping-submit").style.display = "";
          });
        }
        document.querySelector("#grouping-form").addEventListener("submit", (e) => {
          e.preventDefault();
          const choice = document.querySelector('input[name="group"]:checked').value;
          ws.send(JSON.stringify({ kind: "Connect", payload: choice }));
        });
        showOnly("grouping-container");
      } else if (kind === "JoiningSize" && !state.starting) {
        document.title = `Joining [${payload}]`;
        let sizes = document.querySelectorAll(".joining-size");
        for (let s of sizes) {
          s.innerHTML = payload;
        }
        showOnly("waiting-container");
      } else if (kind === "Starting") {
        document.title = "Starting...";
        state.starting = true;
        showOnly("joining-container");
        setTimeout(() => {
          document.location.href = payload;
        }, 3e3);
      } else if (kind === "Instructions") {
        document.querySelector("#instructions-container div").innerHTML = payload;
        show("instructions-container");
      } else if (kind === "Pending") {
        document.title = "Waiting";
        showOnly("pending-container");
      } else if (kind === "Disconnect" && payload.startsWith("Redirect")) {
        const target = payload.replace("Redirect:", "");
        document.location.href = target;
      } else if (kind === "Disconnect" && payload == "Full") {
        showOnly("full-container");
        ws.close();
      } else if (kind === "State" && payload == "Unavailable") {
        document.title = "Unavailable Experiment";
        showOnly("unavailable-container");
        ws.close();
      } else if (kind === "State" && payload == "LandingFailed") {
        showOnly("landing-failed-container");
        ws.close();
      } else if (kind === "Disconnect") {
        ws.close();
      }
    };
  };
  document.addEventListener("DOMContentLoaded", async () => {
    console.log("[supervise] version 0.3.0");
    const slugMatch = /join\/(.*)$/.exec(window.location.pathname);
    if (slugMatch) {
      const slug = slugMatch[1];
      start(slug);
    }
  });
})();
