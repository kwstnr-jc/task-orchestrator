import { useEffect, useState } from "react";
import type { UserInfo } from "./types/task";
import { getMe } from "./lib/api";
import LoginPage from "./pages/LoginPage";
import BoardPage from "./pages/BoardPage";

export default function App() {
  const [user, setUser] = useState<UserInfo | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    getMe()
      .then(setUser)
      .catch(() => setUser(null))
      .finally(() => setLoading(false));
  }, []);

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-gray-400">Loading...</div>
      </div>
    );
  }

  if (!user) {
    return <LoginPage />;
  }

  return <BoardPage user={user} onLogout={() => setUser(null)} />;
}
