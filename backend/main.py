from fastapi import FastAPI, HTTPException, Depends, Header
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
from typing import Optional, List
import sqlite3
import uuid
import datetime
import secrets
import os

app = FastAPI(title="VPN One-Click API", version="1.0.0")

# Enable CORS
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Database setup
DB_PATH = os.getenv("DB_PATH", "vpn_client.db")


def get_db():
    conn = sqlite3.connect(DB_PATH)
    conn.row_factory = sqlite3.Row
    return conn


def init_db():
    conn = get_db()
    cursor = conn.cursor()

    # Devices table
    cursor.execute("""
        CREATE TABLE IF NOT EXISTS devices (
            device_id TEXT PRIMARY KEY,
            device_token TEXT UNIQUE NOT NULL,
            client_version TEXT,
            os_version TEXT,
            hostname TEXT,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            last_seen TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            connected BOOLEAN DEFAULT 0,
            current_node TEXT,
            last_error TEXT
        )
    """)

    # Commands table
    cursor.execute("""
        CREATE TABLE IF NOT EXISTS commands (
            id TEXT PRIMARY KEY,
            device_id TEXT NOT NULL,
            command_type TEXT NOT NULL,
            payload TEXT,
            status TEXT DEFAULT 'pending',
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            executed_at TIMESTAMP,
            FOREIGN KEY (device_id) REFERENCES devices(device_id)
        )
    """)

    # Proxy nodes table
    cursor.execute("""
        CREATE TABLE IF NOT EXISTS proxy_nodes (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            endpoint TEXT NOT NULL,
            protocol TEXT NOT NULL,
            region TEXT,
            active BOOLEAN DEFAULT 1,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
    """)

    # Insert default proxy node
    cursor.execute("""
        INSERT OR IGNORE INTO proxy_nodes (id, endpoint, protocol, region)
        VALUES (1, 'proxy.example.com:443', 'http', 'default')
    """)

    conn.commit()
    conn.close()


# Initialize database on startup
@app.on_event("startup")
async def startup_event():
    init_db()


# Models
class EnrollRequest(BaseModel):
    device_id: str
    client_version: str
    os_version: Optional[str] = None
    hostname: Optional[str] = None


class ProxyProfile(BaseModel):
    node_endpoint: str
    protocol: str
    auth: dict
    expires_at: str


class EnrollResponse(BaseModel):
    device_token: str
    proxy_profile: ProxyProfile


class HeartbeatRequest(BaseModel):
    connected: bool
    current_node: str
    last_error: Optional[str] = None
    client_version: str
    uptime_sec: int


class Command(BaseModel):
    id: str
    type: str
    payload: Optional[dict] = None


class CommandsResponse(BaseModel):
    commands: List[Command]


class CommandResultRequest(BaseModel):
    command_id: str
    status: str
    details: Optional[str] = None


# Dependency for authentication
async def get_device_token(authorization: Optional[str] = Header(None)):
    if not authorization:
        raise HTTPException(status_code=401, detail="Missing authorization header")

    if not authorization.startswith("Bearer "):
        raise HTTPException(status_code=401, detail="Invalid authorization header")

    token = authorization[7:]

    conn = get_db()
    cursor = conn.cursor()
    cursor.execute("SELECT device_id FROM devices WHERE device_token = ?", (token,))
    row = cursor.fetchone()
    conn.close()

    if not row:
        raise HTTPException(status_code=401, detail="Invalid device token")

    return row["device_id"]


# Endpoints
@app.post("/v1/enroll", response_model=EnrollResponse)
async def enroll(request: EnrollRequest):
    """Enroll a new device"""
    conn = get_db()
    cursor = conn.cursor()

    # Check if device already exists
    cursor.execute("SELECT device_token FROM devices WHERE device_id = ?", (request.device_id,))
    existing = cursor.fetchone()

    if existing:
        device_token = existing["device_token"]
    else:
        # Generate new device token
        device_token = secrets.token_urlsafe(32)

        # Insert device
        cursor.execute("""
            INSERT INTO devices (device_id, device_token, client_version, os_version, hostname)
            VALUES (?, ?, ?, ?, ?)
        """, (request.device_id, device_token, request.client_version, request.os_version, request.hostname))

        conn.commit()

    conn.close()

    # Get proxy profile
    proxy_profile = ProxyProfile(
        node_endpoint="proxy.example.com:443",
        protocol="http",
        auth={
            "type": "token",
            "token": secrets.token_urlsafe(16)
        },
        expires_at=(datetime.datetime.utcnow() + datetime.timedelta(hours=24)).isoformat() + "Z"
    )

    return EnrollResponse(
        device_token=device_token,
        proxy_profile=proxy_profile
    )


@app.post("/v1/heartbeat")
async def heartbeat(request: HeartbeatRequest, device_id: str = Depends(get_device_token)):
    """Update device heartbeat"""
    conn = get_db()
    cursor = conn.cursor()

    cursor.execute("""
        UPDATE devices
        SET last_seen = CURRENT_TIMESTAMP,
            connected = ?,
            current_node = ?,
            last_error = ?
        WHERE device_id = ?
    """, (request.connected, request.current_node, request.last_error, device_id))

    conn.commit()
    conn.close()

    return {"status": "ok"}


@app.get("/v1/commands", response_model=CommandsResponse)
async def get_commands(device_id: str = Depends(get_device_token)):
    """Get pending commands for device"""
    conn = get_db()
    cursor = conn.cursor()

    cursor.execute("""
        SELECT id, command_type as type, payload
        FROM commands
        WHERE device_id = ? AND status = 'pending'
        ORDER BY created_at ASC
    """, (device_id,))

    commands = []
    for row in cursor.fetchall():
        commands.append(Command(
            id=row["id"],
            type=row["type"],
            payload=eval(row["payload"]) if row["payload"] else None
        ))

    conn.close()

    return CommandsResponse(commands=commands)


@app.post("/v1/command_result")
async def command_result(request: CommandResultRequest, device_id: str = Depends(get_device_token)):
    """Update command execution result"""
    conn = get_db()
    cursor = conn.cursor()

    cursor.execute("""
        UPDATE commands
        SET status = ?,
            executed_at = CURRENT_TIMESTAMP
        WHERE id = ? AND device_id = ?
    """, (request.status, request.command_id, device_id))

    conn.commit()
    conn.close()

    return {"status": "ok"}


# Admin endpoints (for testing)
@app.get("/admin/devices")
async def list_devices():
    """List all devices (admin)"""
    conn = get_db()
    cursor = conn.cursor()
    cursor.execute("""
        SELECT device_id, client_version, hostname, connected, current_node, last_seen
        FROM devices
        ORDER BY last_seen DESC
    """)

    devices = []
    for row in cursor.fetchall():
        devices.append(dict(row))

    conn.close()
    return {"devices": devices}


@app.post("/admin/commands")
async def create_command(device_id: str, command_type: str, payload: Optional[dict] = None):
    """Create a command for a device (admin)"""
    conn = get_db()
    cursor = conn.cursor()

    command_id = str(uuid.uuid4())

    cursor.execute("""
        INSERT INTO commands (id, device_id, command_type, payload)
        VALUES (?, ?, ?, ?)
    """, (command_id, device_id, command_type, str(payload) if payload else None))

    conn.commit()
    conn.close()

    return {"command_id": command_id, "status": "created"}


@app.get("/")
async def root():
    return {
        "name": "VPN One-Click API",
        "version": "1.0.0",
        "status": "running"
    }


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
