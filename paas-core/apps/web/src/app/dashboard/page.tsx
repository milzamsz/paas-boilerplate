"use client";

import { useEffect, useState } from "react";

interface Org {
    id: string;
    name: string;
    slug: string;
    role: string;
}

export default function DashboardPage() {
    const [orgs, setOrgs] = useState<Org[]>([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const token = localStorage.getItem("token");
        if (!token) {
            window.location.href = "/login";
            return;
        }

        fetch(`${process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080"}/api/v1/orgs`, {
            headers: { Authorization: `Bearer ${token}` },
            credentials: "include",
        })
            .then((r) => r.json())
            .then((data) => {
                setOrgs(Array.isArray(data.data) ? data.data : data.data ? [data.data] : []);
            })
            .catch(() => {
                localStorage.removeItem("token");
                window.location.href = "/login";
            })
            .finally(() => setLoading(false));
    }, []);

    const roleColors: Record<string, string> = {
        owner: "#6366f1",
        admin: "#8b5cf6",
        developer: "#10b981",
        viewer: "#6b7280",
    };

    return (
        <div style={{ minHeight: "100vh" }}>
            {/* Header */}
            <header
                style={{
                    borderBottom: "1px solid #1f2937",
                    padding: "1rem 1.5rem",
                    display: "flex",
                    justifyContent: "space-between",
                    alignItems: "center",
                }}
            >
                <h1 style={{ fontSize: "1.25rem", fontWeight: 700 }}>
                    <span
                        style={{
                            background: "linear-gradient(135deg, #6366f1, #a855f7)",
                            WebkitBackgroundClip: "text",
                            WebkitTextFillColor: "transparent",
                        }}
                    >
                        MyPaaS
                    </span>
                </h1>
                <button
                    className="btn btn-ghost"
                    onClick={() => {
                        localStorage.removeItem("token");
                        window.location.href = "/login";
                    }}
                >
                    Sign out
                </button>
            </header>

            {/* Content */}
            <div className="container" style={{ paddingTop: "2rem" }}>
                <div
                    style={{
                        display: "flex",
                        justifyContent: "space-between",
                        alignItems: "center",
                        marginBottom: "1.5rem",
                    }}
                >
                    <h2 style={{ fontSize: "1.5rem", fontWeight: 600 }}>Organizations</h2>
                    <button className="btn btn-primary">+ New org</button>
                </div>

                {loading ? (
                    <div style={{ textAlign: "center", padding: "4rem", color: "#6b7280" }}>Loading...</div>
                ) : orgs.length === 0 ? (
                    <div className="card" style={{ textAlign: "center", padding: "3rem" }}>
                        <p style={{ fontSize: "1.125rem", fontWeight: 500, marginBottom: "0.5rem" }}>
                            No organizations yet
                        </p>
                        <p style={{ color: "#6b7280", marginBottom: "1.5rem" }}>
                            Create your first organization to start deploying projects.
                        </p>
                        <button className="btn btn-primary">Create organization</button>
                    </div>
                ) : (
                    <div
                        style={{
                            display: "grid",
                            gridTemplateColumns: "repeat(auto-fill, minmax(300px, 1fr))",
                            gap: "1rem",
                        }}
                    >
                        {orgs.map((org) => (
                            <a
                                key={org.id}
                                href={`/dashboard/${org.slug}`}
                                className="card"
                                style={{
                                    cursor: "pointer",
                                    transition: "border-color 0.15s",
                                }}
                            >
                                <div
                                    style={{
                                        display: "flex",
                                        justifyContent: "space-between",
                                        alignItems: "flex-start",
                                    }}
                                >
                                    <div>
                                        <h3 style={{ fontWeight: 600, fontSize: "1.125rem" }}>{org.name}</h3>
                                        <p style={{ color: "#6b7280", fontSize: "0.875rem" }}>/{org.slug}</p>
                                    </div>
                                    <span
                                        style={{
                                            padding: "0.25rem 0.625rem",
                                            borderRadius: "9999px",
                                            fontSize: "0.75rem",
                                            fontWeight: 500,
                                            background: `${roleColors[org.role] || "#6b7280"}20`,
                                            color: roleColors[org.role] || "#6b7280",
                                        }}
                                    >
                                        {org.role}
                                    </span>
                                </div>
                            </a>
                        ))}
                    </div>
                )}
            </div>
        </div>
    );
}
