from __future__ import annotations

from fastapi import APIRouter, Request
from pydantic import BaseModel

router = APIRouter()


class ChatRequest(BaseModel):
    session_id: str
    message: str


class ChatResponse(BaseModel):
    session_id: str
    reply: str
    intent: str
    escalate: bool = False


@router.get("/healthz")
async def health():
    return {"status": "ok"}


@router.post("/chat", response_model=ChatResponse)
async def chat(payload: ChatRequest, request: Request):
    """
    Process an inbound chat message, classify intent, and return a reply.
    Conversation state is persisted in Redis using session_id as key.
    """
    intent = _classify_intent(payload.message)
    reply, escalate = _generate_reply(intent, payload.message)

    # Persist last intent to Redis if available
    redis = getattr(request.app.state, "redis", None)
    if redis is not None:
        await redis.setex(
            f"chatbot:session:{payload.session_id}:last_intent",
            1800,
            intent,
        )

    return ChatResponse(
        session_id=payload.session_id,
        reply=reply,
        intent=intent,
        escalate=escalate,
    )


# ---------------------------------------------------------------------------
# Simple rule-based intent classification
# ---------------------------------------------------------------------------

_INTENT_RULES = [
    (["order status", "where is my order", "track order", "tracking"], "order_status"),
    (["return", "refund", "send back", "exchange"], "return_refund"),
    (["cancel", "cancellation"], "order_cancel"),
    (["payment", "charge", "invoice", "receipt"], "payment_query"),
    (["delivery", "shipping", "ship", "arrive"], "shipping_query"),
    (["password", "login", "account", "sign in"], "account_support"),
    (["hello", "hi", "hey", "help", "start"], "greeting"),
]

_REPLIES = {
    "order_status": (
        "I can help with order status! Please share your order number and "
        "I'll look it up for you.",
        False,
    ),
    "return_refund": (
        "To start a return or refund, please visit your order history and click "
        "'Return Item'. Refunds are processed within 5–7 business days.",
        False,
    ),
    "order_cancel": (
        "Orders can be cancelled within 1 hour of placement. Please check your "
        "order history or contact our support team immediately.",
        False,
    ),
    "payment_query": (
        "For payment-related queries, please check your email for a receipt. "
        "If you believe there is an error, I'll connect you with a specialist.",
        True,
    ),
    "shipping_query": (
        "Standard shipping takes 3–5 business days. Express shipping takes 1–2 days. "
        "You can track your order using the link in your confirmation email.",
        False,
    ),
    "account_support": (
        "For account access issues, please use the 'Forgot Password' link on the "
        "login page. If you still need help, I'll escalate to our support team.",
        True,
    ),
    "greeting": (
        "Hello! I'm the ShopOS support assistant. I can help with order status, "
        "returns, shipping, and account queries. How can I help you today?",
        False,
    ),
    "unknown": (
        "I'm not sure I understood that. Could you rephrase? I can help with "
        "order status, returns, shipping, and account queries.",
        False,
    ),
}


def _classify_intent(message: str) -> str:
    text = message.lower()
    for keywords, intent in _INTENT_RULES:
        if any(kw in text for kw in keywords):
            return intent
    return "unknown"


def _generate_reply(intent: str, _message: str) -> tuple[str, bool]:
    reply, escalate = _REPLIES.get(intent, _REPLIES["unknown"])
    return reply, escalate
