const form = document.getElementById("settings-form");
const tokenInput = document.getElementById("token");
const guildInput = document.getElementById("guild");
const globalInput = document.getElementById("global");

const toggleToken = document.getElementById("toggle-token");
const submitButton = document.getElementById("submit-button");
const clearButton = document.getElementById("clear-button");
const buttonText = document.getElementById("button-text");
const statusElement = document.getElementById("status");

let hasSavedToken = false;

function updateGuildInput() {
    guildInput.disabled = globalInput.checked;

    if (globalInput.checked) {
        guildInput.value = "";
    }
}

function showStatus(message, type) {
    statusElement.textContent = message;
    statusElement.className = `status ${type}`;
}

function setLoading(loading) {
    submitButton.disabled = loading;
    clearButton.disabled = loading;

    buttonText.textContent = loading
        ? "Saving configuration..."
        : "Save & Start GhostDrop";
}

function updateTokenPlaceholder() {
    tokenInput.placeholder = hasSavedToken
        ? "•••••••• (saved — leave blank to keep)"
        : "Paste your Discord bot token";
}

async function loadSettings() {
    try {
        const response = await fetch("/api/settings", {
            cache: "no-store"
        });

        if (!response.ok) {
            throw new Error("Failed to load settings.");
        }

        const settings = await response.json();

        hasSavedToken = settings.has_token;
        guildInput.value = settings.guild_id || "";
        globalInput.checked = !settings.guild_id;

        updateTokenPlaceholder();
        updateGuildInput();
    } catch (error) {
        showStatus(error.message, "error");
    }
}

toggleToken.addEventListener("click", () => {
    const hidden = tokenInput.type === "password";

    tokenInput.type = hidden ? "text" : "password";
    toggleToken.textContent = hidden ? "Hide" : "Show";
    toggleToken.setAttribute(
        "aria-label",
        hidden ? "Hide token" : "Show token"
    );
});

globalInput.addEventListener("change", updateGuildInput);

guildInput.addEventListener("input", () => {
    guildInput.value = guildInput.value.replace(/\D/g, "");
});

form.addEventListener("submit", async (event) => {
    event.preventDefault();

    const token = tokenInput.value.trim();
    const guildID = globalInput.checked
        ? ""
        : guildInput.value.trim();

    if (!token && !hasSavedToken) {
        showStatus("Please enter your Discord bot token.", "error");
        return;
    }

    if (!globalInput.checked && !/^\d{17,20}$/.test(guildID)) {
        showStatus("Please enter a valid Discord Guild ID.", "error");
        return;
    }

    setLoading(true);
    statusElement.className = "status";
    statusElement.textContent = "";

    try {
        const response = await fetch("/api/settings", {
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify({
                token: token,
                guild_id: guildID
            })
        });

        const result = await response.json();
        if (!response.ok) {
            throw new Error(
                result.error || "Failed to save configuration."
            );
        }

        hasSavedToken = true;
        tokenInput.value = "";
        tokenInput.type = "password";
        toggleToken.textContent = "Show";
        updateTokenPlaceholder();

        showStatus(
            "Configuration saved successfully!",
            "success"
        );
    } catch (error) {
        showStatus(error.message, "error");
    } finally {
        setLoading(false);
    }
});

clearButton.addEventListener("click", async () => {
    const confirmed = confirm(
        "Are you sure you want to clear the saved configuration?"
    );

    if (!confirmed) {
        return;
    }

    clearButton.disabled = true;
    submitButton.disabled = true;

    try {
        const response = await fetch("/api/settings", {
            method: "DELETE",
            headers: {
                "Content-Type": "application/json"
            }
        });

        const result = await response.json();
        if (!response.ok) {
            throw new Error(
                result.error || "Failed to clear configuration."
            );
        }

        hasSavedToken = false;

        form.reset();
        tokenInput.value = "";
        tokenInput.type = "password";
        toggleToken.textContent = "Show";

        guildInput.value = "";
        globalInput.checked = true;

        updateTokenPlaceholder();
        updateGuildInput();

        showStatus(
            "Configuration cleared successfully!",
            "success"
        );
    } catch (error) {
        showStatus(error.message, "error");
    } finally {
        clearButton.disabled = false;
        submitButton.disabled = false;
    }
});

updateGuildInput();
loadSettings();
