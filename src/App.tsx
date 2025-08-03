import { useNavigate } from "@solidjs/router";
import { type Component, createSignal } from "solid-js";

const App: Component = () => {
  const navigate = useNavigate();
  const [prompt, setPrompt] = createSignal("");
  const disabled = () => prompt() === "";

  return (
    <form
      class="flex flex-col items-stretch justify-center min-h-screen mx-auto p-8 space-y-4 w-full sm:w-[40rem]"
      onsubmit={handleSubmit}
    >
      <div class="font-bold text-4xl text-center">Need a form?</div>
      <input
        class="border p-2 rounded-lg"
        oninput={(e) => setPrompt(e.target.value)}
        placeholder="Prompt"
        type="text"
        value={prompt()}
      />
      <button
        class="border p-2 rounded-lg transition"
        classList={{
          "opacity-25": disabled(),
          "hover:bg-black/5": !disabled(),
          "active:bg-black/10": !disabled(),
        }}
        disabled={disabled()}
        type="submit"
      >
        Make me one
      </button>
    </form>
  );

  async function handleSubmit(evt: SubmitEvent) {
    evt.preventDefault();
    const res = await fetch("/api/v1/form", {
      body: JSON.stringify({ prompt: prompt() }),
      headers: { "Content-Type": "application/json" },
      method: "POST",
    });
    if (res.status !== 200) {
      // TODO: use solid toast and then set up a toast here.
      return;
    }
    const body = await res.json();
    navigate(`/${body.id}`);
  }
};

export default App;
