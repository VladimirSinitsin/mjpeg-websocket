const API_VERSION = "v1";

const API_BASE =
    (window.STREAM_API || `${window.location.protocol}//${window.location.hostname}:8080`) +
    `/${API_VERSION}`;
const WS_BASE = API_BASE.replace(/^http/, "ws");

const streamList = document.getElementById("stream-list");
const refreshBtn = document.getElementById("refresh");
const startWsBtn = document.getElementById("start-ws");
const stopBtn = document.getElementById("stop");
const currentStreamLabel = document.getElementById("current-stream");
const statusEl = document.getElementById("status");
const mjpegView = document.getElementById("mjpeg-view");
const wsCanvas = document.getElementById("ws-canvas");
const wsCtx = wsCanvas.getContext("2d");

const editBtn = document.getElementById("edit");
const editDialog = document.getElementById("edit-dialog");
const editForm = document.getElementById("edit-form");
const editTitle = document.getElementById("edit-title");
const editDescription = document.getElementById("edit-description");
const editFrameInterval = document.getElementById("edit-frame-interval");
const editCancelBtn = document.getElementById("edit-cancel-btn");

let selectedStream = null;
let wsConnection = null;
let currentMjpegUrl = null;

function normalizeStream(stream) {
    return {
        ...stream,
        frame_count: stream.frame_count ?? stream.frameCount,
        frame_interval_ms: stream.frame_interval_ms ?? stream.frameIntervalMs,
        created_at: stream.created_at ?? stream.createdAt,
        updated_at: stream.updated_at ?? stream.updatedAt,
    };
}

function logStatus(msg) {
    const now = new Date().toLocaleTimeString();
    statusEl.textContent = `[${now}] ${msg}\n` + statusEl.textContent;
}

function revokeCurrentMjpegUrl() {
    if (currentMjpegUrl) {
        URL.revokeObjectURL(currentMjpegUrl);
        currentMjpegUrl = null;
    }
}

async function fetchStreams() {
    try {
        const res = await fetch(`${API_BASE}/streams`);
        if (!res.ok) throw new Error(`HTTP ${res.status}`);
        const data = await res.json();
        const streams = (data.streams || []).map(normalizeStream);
        renderStreams(streams);
    } catch (err) {
        logStatus(`Error loading streams: ${err.message}`);
    }
}

function renderStreams(streams) {
    streamList.innerHTML = "";
    streams.forEach((stream) => {
        const li = document.createElement("li");
        li.textContent = `${stream.title} (${stream.frame_count ?? "?"} frames)`;
        li.dataset.id = stream.id;
        li.addEventListener("click", () => selectStream(stream));
        if (selectedStream && selectedStream.id === stream.id) {
            li.classList.add("active");
        }
        streamList.appendChild(li);
    });
    if (!streams.length) {
        const empty = document.createElement("li");
        empty.textContent = "No streams";
        empty.classList.add("disabled");
        streamList.appendChild(empty);
    }
}

function selectStream(stream) {
    selectedStream = stream;
    Array.from(streamList.children).forEach((child) => {
        child.classList.toggle("active", child.dataset.id === stream.id);
    });
    currentStreamLabel.textContent = `Stream: ${stream.title}`;
    startWsBtn.disabled = false;
    stopBtn.disabled = false;
    editBtn.disabled = false;
}

function startWebSocket() {
    if (!selectedStream) return;
    cleanupWs();
    revokeCurrentMjpegUrl();
    mjpegView.src = "";
    wsCanvas.hidden = false;
    mjpegView.hidden = true;
    wsCanvas.width = 640;
    wsCanvas.height = 360;

    const ws = new WebSocket(`${WS_BASE}/streams/${selectedStream.id}/ws`);
    ws.binaryType = "arraybuffer";

    ws.onopen = () => {
        logStatus("WebSocket connected");
    };

    ws.onmessage = (event) => {
        if (typeof event.data === "string") {
            try {
                const meta = JSON.parse(event.data);
                logStatus(`Frame ${meta.sequence} (${meta.mime_type})`);
            } catch (err) {
                logStatus(`Error parsing metadata: ${err.message}`);
            }
            return;
        }

        const blob = new Blob([event.data]);
        const img = new Image();
        const url = URL.createObjectURL(blob);
        img.onload = () => {
            wsCtx.clearRect(0, 0, wsCanvas.width, wsCanvas.height);
            wsCtx.drawImage(img, 0, 0, wsCanvas.width, wsCanvas.height);
            URL.revokeObjectURL(url);
        };
        img.src = url;
    };

    ws.onerror = (event) => {
        console.error("WS error", event);
        logStatus("WebSocket error");
    };

    ws.onclose = (event) => {
        logStatus(`WebSocket closed: ${event.reason || event.code}`);
        cleanupWs();
    };

    wsConnection = ws;
}

function cleanupWs() {
    if (wsConnection) {
        wsConnection.onopen = null;
        wsConnection.onmessage = null;
        wsConnection.onerror = null;
        wsConnection.onclose = null;
        if (
            wsConnection.readyState === WebSocket.OPEN ||
            wsConnection.readyState === WebSocket.CONNECTING
        ) {
            wsConnection.close();
        }
        wsConnection = null;
    }
}

function stopStreaming() {
    cleanupWs();
    revokeCurrentMjpegUrl();
    mjpegView.src = "";
    mjpegView.hidden = true;
    wsCanvas.hidden = true;
    stopBtn.disabled = true;
    startWsBtn.disabled = !selectedStream;
    editBtn.disabled = !selectedStream;
    logStatus("Stopped");
}

function openEditModal() {
    if (!selectedStream) return;
    editTitle.value = selectedStream.title || "";
    editDescription.value = selectedStream.description || "";
    editFrameInterval.value =
        selectedStream.frame_interval_ms ?? selectedStream.frameIntervalMs ?? "";
    editDialog.showModal();
}

async function updateStream(event) {
    event.preventDefault();
    if (!selectedStream) { editDialog.close(); return; }

    editDialog.close();

    const payload = {
        id: selectedStream.id,
        title: editTitle.value,
        description: editDescription.value,
    };

    const intervalVal = String(editFrameInterval.value || "").trim();
    if (intervalVal !== "") payload.frameIntervalMs = Number(intervalVal);

    try {
        const res = await fetch(`${API_BASE}/streams/${selectedStream.id}`, {
            method: "PUT",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(payload),
        });
        if (!res.ok) {
            const err = await res.json().catch(() => ({}));
            throw new Error(err.error || `HTTP ${res.status}`);
        }

        const data = await res.json().catch(() => ({}));
        const updated = data.stream ? normalizeStream(data.stream) : normalizeStream(data);

        if (updated && updated.id === selectedStream.id) {
            selectedStream = updated;
        }

        logStatus("Stream updated");
        await fetchStreams();

        if (selectedStream) {
            const li = Array.from(streamList.children).find(
                (el) => el.dataset.id === selectedStream.id
            );
            if (li) li.classList.add("active");
            currentStreamLabel.textContent = `Stream: ${selectedStream.title}`;
        }
    } catch (err) {
        logStatus(`Error updating stream: ${err.message}`);
    }
}

function cancelEdit() {
    editDialog.close();
}

refreshBtn.addEventListener("click", fetchStreams);
startWsBtn.addEventListener("click", startWebSocket);
stopBtn.addEventListener("click", stopStreaming);
editBtn.addEventListener("click", openEditModal);
editForm.addEventListener("submit", updateStream);
editCancelBtn.addEventListener("click", cancelEdit);

fetchStreams();
