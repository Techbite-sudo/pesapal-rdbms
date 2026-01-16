"use client";

import { useState, useEffect } from "react";

interface QueryResponse {
  success: boolean;
  message?: string;
  columns?: string[];
  rows?: any[][];
  rowsAffected: number;
  error?: string;
}

interface TableInfo {
  name: string;
  columns: {
    name: string;
    dataType: string;
    size?: number;
    primaryKey: boolean;
    unique: boolean;
    notNull: boolean;
  }[];
}

export default function Home() {
  const [query, setQuery] = useState("");
  const [result, setResult] = useState<QueryResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [tables, setTables] = useState<TableInfo[]>([]);
  const [apiUrl] = useState("http://localhost:8080");

  useEffect(() => {
    fetchTables();
  }, []);

  const fetchTables = async () => {
    try {
      const response = await fetch(`${apiUrl}/api/tables`);
      const data = await response.json();
      if (data.success) {
        setTables(data.tables);
      }
    } catch (error) {
      console.error("Failed to fetch tables:", error);
    }
  };

  const executeQuery = async () => {
    if (!query.trim()) return;

    setLoading(true);
    setResult(null);

    try {
      const response = await fetch(`${apiUrl}/api/query`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ query }),
      });

      const data = await response.json();
      setResult(data);

      // Refresh tables list if it was a DDL operation
      if (
        query.toUpperCase().includes("CREATE TABLE") ||
        query.toUpperCase().includes("DROP TABLE")
      ) {
        fetchTables();
      }
    } catch (error) {
      setResult({
        success: false,
        error: `Network error: ${error}`,
        rowsAffected: 0,
      });
    } finally {
      setLoading(false);
    }
  };

  const loadExample = (example: string) => {
    setQuery(example);
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100">
      <div className="container mx-auto px-4 py-8">
        {/* Header */}
        <div className="bg-white rounded-lg shadow-lg p-6 mb-6">
          <h1 className="text-4xl font-bold text-indigo-600 mb-2">
            Pesapal RDBMS
          </h1>
          <p className="text-gray-600">
            Junior Dev Challenge 2026 - Interactive Database Management System
          </p>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Left Column - Query Editor */}
          <div className="lg:col-span-2 space-y-6">
            {/* Query Editor */}
            <div className="bg-white rounded-lg shadow-lg p-6">
              <h2 className="text-2xl font-semibold text-gray-800 mb-4">
                SQL Query Editor
              </h2>
              <textarea
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                placeholder="Enter your SQL query here..."
                className="w-full h-40 p-4 border border-gray-300 rounded-lg font-mono text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
              />
              <div className="flex gap-2 mt-4">
                <button
                  onClick={executeQuery}
                  disabled={loading || !query.trim()}
                  className="px-6 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 disabled:bg-gray-400 disabled:cursor-not-allowed transition-colors"
                >
                  {loading ? "Executing..." : "Execute Query"}
                </button>
                <button
                  onClick={() => setQuery("")}
                  className="px-6 py-2 bg-gray-200 text-gray-700 rounded-lg hover:bg-gray-300 transition-colors"
                >
                  Clear
                </button>
              </div>
            </div>

            {/* Results */}
            {result && (
              <div className="bg-white rounded-lg shadow-lg p-6">
                <h2 className="text-2xl font-semibold text-gray-800 mb-4">
                  Results
                </h2>
                {result.success ? (
                  <>
                    {result.message && (
                      <div className="bg-green-50 border border-green-200 text-green-800 px-4 py-3 rounded-lg mb-4">
                        {result.message}
                      </div>
                    )}
                    {result.columns && result.rows && (
                      <div className="overflow-x-auto">
                        <table className="min-w-full divide-y divide-gray-200">
                          <thead className="bg-gray-50">
                            <tr>
                              {result.columns.map((col, idx) => (
                                <th
                                  key={idx}
                                  className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                                >
                                  {col}
                                </th>
                              ))}
                            </tr>
                          </thead>
                          <tbody className="bg-white divide-y divide-gray-200">
                            {result.rows.map((row, rowIdx) => (
                              <tr key={rowIdx}>
                                {row.map((cell, cellIdx) => (
                                  <td
                                    key={cellIdx}
                                    className="px-6 py-4 whitespace-nowrap text-sm text-gray-900"
                                  >
                                    {cell === null ? (
                                      <span className="text-gray-400 italic">
                                        NULL
                                      </span>
                                    ) : (
                                      String(cell)
                                    )}
                                  </td>
                                ))}
                              </tr>
                            ))}
                          </tbody>
                        </table>
                        <p className="text-sm text-gray-600 mt-4">
                          {result.rowsAffected} row(s) returned
                        </p>
                      </div>
                    )}
                  </>
                ) : (
                  <div className="bg-red-50 border border-red-200 text-red-800 px-4 py-3 rounded-lg">
                    <strong>Error:</strong> {result.error}
                  </div>
                )}
              </div>
            )}
          </div>

          {/* Right Column - Tables & Examples */}
          <div className="space-y-6">
            {/* Tables */}
            <div className="bg-white rounded-lg shadow-lg p-6">
              <h2 className="text-2xl font-semibold text-gray-800 mb-4">
                Tables
              </h2>
              {tables.length === 0 ? (
                <p className="text-gray-500 italic">No tables yet</p>
              ) : (
                <div className="space-y-3">
                  {tables.map((table) => (
                    <div
                      key={table.name}
                      className="border border-gray-200 rounded-lg p-3"
                    >
                      <h3 className="font-semibold text-indigo-600 mb-2">
                        {table.name}
                      </h3>
                      <div className="text-xs space-y-1">
                        {table.columns.map((col) => (
                          <div key={col.name} className="text-gray-600">
                            <span className="font-mono">{col.name}</span>
                            <span className="text-gray-400 ml-2">
                              {col.dataType}
                              {col.size ? `(${col.size})` : ""}
                            </span>
                            {col.primaryKey && (
                              <span className="ml-2 text-xs bg-blue-100 text-blue-800 px-1 rounded">
                                PK
                              </span>
                            )}
                            {col.unique && (
                              <span className="ml-1 text-xs bg-purple-100 text-purple-800 px-1 rounded">
                                UNIQUE
                              </span>
                            )}
                          </div>
                        ))}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>

            {/* Examples */}
            <div className="bg-white rounded-lg shadow-lg p-6">
              <h2 className="text-2xl font-semibold text-gray-800 mb-4">
                Examples
              </h2>
              <div className="space-y-2">
                <button
                  onClick={() =>
                    loadExample(
                      "CREATE TABLE users (id INTEGER PRIMARY KEY, name VARCHAR(100), email VARCHAR(100) UNIQUE);"
                    )
                  }
                  className="w-full text-left px-3 py-2 text-sm bg-gray-50 hover:bg-gray-100 rounded border border-gray-200 transition-colors"
                >
                  Create Table
                </button>
                <button
                  onClick={() =>
                    loadExample(
                      "INSERT INTO users VALUES (1, 'John Doe', 'john@example.com');"
                    )
                  }
                  className="w-full text-left px-3 py-2 text-sm bg-gray-50 hover:bg-gray-100 rounded border border-gray-200 transition-colors"
                >
                  Insert Data
                </button>
                <button
                  onClick={() => loadExample("SELECT * FROM users;")}
                  className="w-full text-left px-3 py-2 text-sm bg-gray-50 hover:bg-gray-100 rounded border border-gray-200 transition-colors"
                >
                  Select All
                </button>
                <button
                  onClick={() =>
                    loadExample(
                      "UPDATE users SET name = 'Jane Doe' WHERE id = 1;"
                    )
                  }
                  className="w-full text-left px-3 py-2 text-sm bg-gray-50 hover:bg-gray-100 rounded border border-gray-200 transition-colors"
                >
                  Update Data
                </button>
                <button
                  onClick={() => loadExample("DELETE FROM users WHERE id = 1;")}
                  className="w-full text-left px-3 py-2 text-sm bg-gray-50 hover:bg-gray-100 rounded border border-gray-200 transition-colors"
                >
                  Delete Data
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
