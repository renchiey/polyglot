import { BrowserRouter, Route, Routes } from "react-router-dom";
import { PlugZap, Languages } from "lucide-react";
import { SessionProvider, useSession } from "./lib/session";
import { SettingsProvider } from "./lib/settings";
import { ToastHost, Button, Card, Spinner } from "./components/ui";
import { AppShell } from "./components/AppShell";
import { DashboardView } from "./features/DashboardView";
import { JourneyView } from "./features/JourneyView";
import { ReadView } from "./features/ReadView";
import { SettingsView } from "./features/SettingsView";
import { StudyView } from "./features/StudyView";
import { VaultView } from "./features/VaultView";

export default function App() {
  return (
    <SessionProvider>
      <SettingsProvider>
        <ToastHost>
          <BrowserRouter>
            <Shell />
          </BrowserRouter>
        </ToastHost>
      </SettingsProvider>
    </SessionProvider>
  );
}

function Shell() {
  const { status } = useSession();
  return <AppShell>{status === "ready" ? <AppRoutes /> : <Gate />}</AppShell>;
}

function AppRoutes() {
  return (
    <Routes>
      <Route path="/" element={<DashboardView />} />
      <Route path="/journey" element={<JourneyView />} />
      <Route path="/read" element={<ReadView />} />
      <Route path="/study" element={<StudyView />} />
      <Route path="/vault" element={<VaultView />} />
      <Route path="/settings" element={<SettingsView />} />
      <Route path="*" element={<DashboardView />} />
    </Routes>
  );
}

function Gate() {
  const { status, error, retry } = useSession();

  if (status === "loading") {
    return (
      <Card className="mx-auto mt-12 max-w-md">
        <div className="flex flex-col items-center gap-4 py-8 text-center">
          <div className="animate-float text-[var(--color-grape)]"><Languages size={40} /></div>
          <Spinner />
          <p className="font-display font-bold text-[var(--color-ink-soft)]">Waking up the server…</p>
        </div>
      </Card>
    );
  }

  return (
    <Card className="mx-auto mt-12 max-w-md">
      <div className="flex flex-col items-center gap-3 py-6 text-center">
        <div className="text-[var(--color-danger)]"><PlugZap size={40} /></div>
        <h2 className="text-xl">Can't reach the API</h2>
        <p className="text-sm text-[var(--color-ink-soft)]">
          {error || "The Go server isn't responding."} Start it with{" "}
          <code className="rounded bg-white/70 px-1.5 py-0.5 font-mono text-xs">make server</code>, and set{" "}
          <code className="rounded bg-white/70 px-1.5 py-0.5 font-mono text-xs">VITE_API_URL</code> in{" "}
          <code className="rounded bg-white/70 px-1.5 py-0.5 font-mono text-xs">web/.env</code> if it isn't on :8080.
        </p>
        <Button accent="grape" onClick={retry}>Try again</Button>
      </div>
    </Card>
  );
}
