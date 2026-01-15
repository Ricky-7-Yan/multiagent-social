import React, { useEffect, useState, useRef } from "react";

type Agent = {
  id: string;
  name: string;
  persona: string;
};

type Conversation = {
  id: string;
  title?: string;
};

export default function App() {
  const [agents, setAgents] = useState<Agent[]>([]);
  const [convs, setConvs] = useState<Conversation[]>([]);
  const [name, setName] = useState("");
  const [persona, setPersona] = useState("");
  const [convId, setConvId] = useState("");
  const [messages, setMessages] = useState<any[]>([]);
  const wsRef = useRef<WebSocket | null>(null);
  const msgRef = useRef<HTMLInputElement | null>(null);

  useEffect(() => {
    listAgents();
    listConversations();
  }, []);

  async function listAgents() {
    const res = await fetch("/api/v1/agents");
    const j = await res.json();
    setAgents(j);
  }

  async function createAgent() {
    await fetch("/api/v1/agents", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name, persona }),
    });
    setName("");
    setPersona("");
    listAgents();
  }

  async function createConversation() {
    const res = await fetch("/api/v1/conversations", { method: "POST" });
    const id = await res.text();
    setConvId(id);
    listConversations();
  }

  async function listConversations() {
    const res = await fetch("/api/v1/conversations");
    const j = await res.json();
    setConvs(j);
  }

  function openWs() {
    if (!convId) return alert("missing id");
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }
    const w = new WebSocket(`ws://${location.host}/ws/conversations/${convId}`);
    w.onmessage = (ev) => {
      try {
        const j = JSON.parse(ev.data);
        setMessages((m) => [...m, j]);
      } catch {
        setMessages((m) => [...m, { raw: ev.data }]);
      }
    };
    wsRef.current = w;
  }

  async function postMessage() {
    if (!convId) return alert("missing id");
    const text = msgRef.current?.value || "";
    if (!text) return;
    await fetch(`/api/v1/conversations/${convId}/messages`, { method: "POST", body: text });
    if (msgRef.current) msgRef.current.value = "";
  }

  return (
    <div className="container">
      <h1>MultiAgent Admin (React)</h1>
      <section>
        <h2>Create Agent</h2>
        <input value={name} onChange={(e) => setName(e.target.value)} placeholder="Agent name" />
        <textarea value={persona} onChange={(e) => setPersona(e.target.value)} placeholder="Persona" />
        <button onClick={createAgent}>Create Agent</button>
      </section>
      <section>
        <h2>Agents</h2>
        <button onClick={listAgents}>Refresh</button>
        <ul>
          {agents.map((a) => (
            <li key={a.id}>
              {a.name} - <code>{a.id}</code>
            </li>
          ))}
        </ul>
      </section>
      <section>
        <h2>Conversations</h2>
        <button onClick={createConversation}>Create Conversation</button>
        <button onClick={listConversations}>Refresh</button>
        <ul>
          {convs.map((c) => (
            <li key={c.id}>
              <b>{c.title || "untitled"}</b> - <code>{c.id}</code>{" "}
              <button onClick={() => setConvId(c.id)}>Select</button>
            </li>
          ))}
        </ul>
      </section>
      <section>
        <h2>Conversation Viewer</h2>
        <div className="row">
          <input value={convId} onChange={(e) => setConvId(e.target.value)} placeholder="conversation id" />
          <button onClick={openWs}>Open WS</button>
        </div>
        <div className="messages">
          {messages.map((m, i) => (
            <div key={i} className="msg">
              {JSON.stringify(m)}
            </div>
          ))}
        </div>
        <div className="row">
          <input ref={msgRef} placeholder="message" />
          <button onClick={postMessage}>Send</button>
        </div>
      </section>
    </div>
  );
}

