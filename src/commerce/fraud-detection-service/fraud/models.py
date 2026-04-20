from datetime import datetime
from pydantic import BaseModel


class FraudCheckRequest(BaseModel):
    order_id: str
    customer_id: str
    amount: float
    currency: str = "USD"
    ip_address: str = ""
    device_fingerprint: str = ""
    items: list[dict] = []
    shipping_address: dict = {}
    billing_address: dict = {}


class FraudCheckResult(BaseModel):
    order_id: str
    risk_score: int          # 0-100
    risk_level: str          # "low" | "medium" | "high" | "critical"
    decision: str            # "approve" | "review" | "decline"
    signals: list[str]       # human-readable reasons
    checked_at: datetime
