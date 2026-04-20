from __future__ import annotations

import re
from datetime import datetime, timezone
from typing import Optional

from .models import SentimentLabel, SentimentResult

# ---------------------------------------------------------------------------
# Word lists
# ---------------------------------------------------------------------------

POSITIVE_WORDS: frozenset[str] = frozenset(
    {
        "excellent", "amazing", "love", "perfect", "great", "fantastic",
        "wonderful", "awesome", "outstanding", "superb", "brilliant",
        "smooth", "fast", "reliable", "helpful", "satisfied", "pleased",
        "delighted", "recommend", "recommended", "beautiful", "gorgeous",
        "incredible", "impressive", "exceptional", "magnificent", "fabulous",
        "terrific", "marvelous", "phenomenal", "splendid", "glorious",
        "spectacular", "exquisite", "superb", "top-notch", "top-quality",
        "high-quality", "well-made", "durable", "sturdy", "solid", "strong",
        "premium", "luxurious", "elegant", "stylish", "attractive", "lovely",
        "charming", "appealing", "pleasant", "enjoyable", "fun", "exciting",
        "thrilling", "entertaining", "engaging", "interesting", "fascinating",
        "captivating", "compelling", "informative", "educational", "useful",
        "handy", "convenient", "easy", "intuitive", "user-friendly", "simple",
        "clear", "clean", "neat", "tidy", "organised", "organized", "efficient",
        "effective", "productive", "powerful", "innovative", "creative",
        "unique", "original", "fresh", "new", "improved", "better", "best",
        "superior", "advanced", "modern", "cutting-edge", "state-of-the-art",
        "revolutionary", "groundbreaking", "game-changing", "life-changing",
        "transformative", "impressive", "astounding", "astonishing",
        "extraordinary", "remarkable", "notable", "noteworthy", "significant",
        "meaningful", "valuable", "worthwhile", "worth", "affordable",
        "competitive", "fair", "reasonable", "bargain", "deal", "value",
        "quality", "genuine", "authentic", "real", "true", "honest",
        "trustworthy", "dependable", "consistent", "stable", "robust",
        "accurate", "precise", "correct", "right", "perfect", "flawless",
        "pristine", "immaculate", "spotless", "polished", "refined",
        "professional", "competent", "skilled", "knowledgeable", "expert",
        "thorough", "meticulous", "careful", "attentive", "responsive",
        "prompt", "quick", "timely", "punctual", "efficient", "swift",
        "speedy", "rapid", "immediate", "instant", "seamless", "frictionless",
        "happy", "glad", "grateful", "thankful", "appreciative", "content",
        "cheerful", "joyful", "ecstatic", "elated", "overjoyed",
        "comfortable", "cozy", "snug", "warm", "friendly", "kind", "gentle",
        "supportive", "encouraging", "motivating", "inspiring", "uplifting",
        "positive", "optimistic", "hopeful", "confident", "secure", "safe",
    }
)

NEGATIVE_WORDS: frozenset[str] = frozenset(
    {
        "terrible", "awful", "horrible", "poor", "bad", "broken",
        "disappointing", "slow", "defective", "useless", "frustrating",
        "annoying", "waste", "unreliable", "unhappy", "dissatisfied",
        "refund", "return", "cheap", "fake", "counterfeit", "faulty",
        "damaged", "defective", "inferior", "substandard", "low-quality",
        "flimsy", "fragile", "weak", "unstable", "unreliable", "inconsistent",
        "inaccurate", "incorrect", "wrong", "misleading", "deceptive",
        "fraudulent", "scam", "ripoff", "overpriced", "expensive", "costly",
        "worthless", "useless", "pointless", "unnecessary", "redundant",
        "outdated", "obsolete", "old", "dated", "stale", "expired",
        "ugly", "hideous", "disgusting", "revolting", "repulsive", "nasty",
        "dirty", "filthy", "greasy", "sticky", "smelly", "stinky",
        "noisy", "loud", "harsh", "rough", "scratchy", "uncomfortable",
        "painful", "harmful", "dangerous", "risky", "unsafe", "hazardous",
        "toxic", "poisonous", "unhealthy", "sick", "ill", "broken",
        "malfunctioning", "glitchy", "buggy", "laggy", "unresponsive",
        "complicated", "confusing", "unclear", "vague", "ambiguous",
        "misleading", "deceptive", "dishonest", "untrustworthy", "shady",
        "suspicious", "sketchy", "dodgy", "problematic", "troublesome",
        "difficult", "hard", "challenging", "impossible", "impractical",
        "inconvenient", "tedious", "boring", "dull", "uninteresting",
        "bland", "mediocre", "average", "ordinary", "plain", "generic",
        "disappointing", "let-down", "letdown", "underwhelming", "overrated",
        "hyped", "gimmick", "waste", "garbage", "junk", "trash", "rubbish",
        "crap", "pathetic", "laughable", "ridiculous", "absurd", "stupid",
        "idiotic", "nonsensical", "illogical", "inefficient", "ineffective",
        "unproductive", "counterproductive", "negative", "pessimistic",
        "depressing", "sad", "upset", "angry", "furious", "outraged",
        "disgusted", "offended", "insulted", "disrespected", "ignored",
        "neglected", "abandoned", "cheated", "stolen", "robbed", "lied",
        "deceived", "manipulated", "exploited", "abused", "mistreated",
        "rude", "impolite", "unprofessional", "careless", "negligent",
        "irresponsible", "incompetent", "unskilled", "unqualified",
        "failure", "failed", "failing", "unsuccessful", "rejected", "denied",
        "refused", "blocked", "stuck", "frozen", "crashed", "error",
        "broken", "missing", "lost", "delayed", "late", "overdue",
    }
)

NEGATION_WORDS: frozenset[str] = frozenset(
    {
        "not", "no", "never", "neither", "nor",
        "don't", "doesnt", "doesn't", "didnt", "didn't",
        "wont", "won't", "cant", "can't", "couldnt", "couldn't",
        "isnt", "isn't", "wasnt", "wasn't", "arent", "aren't",
        "werent", "weren't", "havent", "haven't", "hadnt", "hadn't",
        "shouldnt", "shouldn't", "wouldnt", "wouldn't",
        "without", "barely", "hardly", "scarcely", "seldom",
    }
)

# Score thresholds
POSITIVE_THRESHOLD = 0.6
NEGATIVE_THRESHOLD = 0.4


def _tokenize(text: str) -> list[str]:
    """Lowercase and split on non-word characters, preserving contractions."""
    # Normalise common contractions before splitting
    text = text.lower()
    text = re.sub(r"n't", " n't", text)
    tokens = re.findall(r"[\w']+(?:-[\w']+)*", text)
    return tokens


class RuleBasedAnalyzer:
    """
    Lexicon-based sentiment analyser with negation window support.

    Negation window: when a negation token is found, the next
    NEGATION_WINDOW tokens have their sentiment flipped.
    """

    NEGATION_WINDOW: int = 3

    def analyze(
        self,
        text: str,
        entity_id: Optional[str] = None,
        entity_type: Optional[str] = None,
    ) -> SentimentResult:
        if not text or not text.strip():
            return SentimentResult(
                text=text,
                label=SentimentLabel.NEUTRAL,
                score=0.5,
                positiveWords=[],
                negativeWords=[],
                entityId=entity_id,
                entityType=entity_type,
                analyzedAt=datetime.now(timezone.utc),
            )

        tokens = _tokenize(text)
        found_positive: list[str] = []
        found_negative: list[str] = []

        negation_remaining = 0

        for token in tokens:
            is_negation = token in NEGATION_WORDS
            if is_negation:
                negation_remaining = self.NEGATION_WINDOW
                continue

            is_positive = token in POSITIVE_WORDS
            is_negative = token in NEGATIVE_WORDS

            if not is_positive and not is_negative:
                if negation_remaining > 0:
                    negation_remaining -= 1
                continue

            # Flip under negation
            if negation_remaining > 0:
                negation_remaining -= 1
                if is_positive:
                    # positive word negated → counts as negative
                    found_negative.append(token)
                elif is_negative:
                    # negative word negated → counts as positive
                    found_positive.append(token)
            else:
                if is_positive:
                    found_positive.append(token)
                if is_negative:
                    found_negative.append(token)

        pos_count = len(found_positive)
        neg_count = len(found_negative)
        total = pos_count + neg_count

        # Score: fraction of sentiment-bearing words that are positive
        # Add +1 smoothing in denominator to avoid division by zero
        score = pos_count / (total + 1)

        if score > POSITIVE_THRESHOLD:
            label = SentimentLabel.POSITIVE
        elif score < NEGATIVE_THRESHOLD:
            label = SentimentLabel.NEGATIVE
        else:
            label = SentimentLabel.NEUTRAL

        return SentimentResult(
            text=text,
            label=label,
            score=round(score, 4),
            positiveWords=found_positive,
            negativeWords=found_negative,
            entityId=entity_id,
            entityType=entity_type,
            analyzedAt=datetime.now(timezone.utc),
        )

    def batch_analyze(
        self,
        texts: list[str],
        entity_ids: Optional[list[Optional[str]]] = None,
        entity_types: Optional[list[Optional[str]]] = None,
    ) -> list[SentimentResult]:
        results: list[SentimentResult] = []
        for i, text in enumerate(texts):
            eid = entity_ids[i] if entity_ids and i < len(entity_ids) else None
            etype = entity_types[i] if entity_types and i < len(entity_types) else None
            results.append(self.analyze(text, entity_id=eid, entity_type=etype))
        return results


# Module-level singleton
analyzer = RuleBasedAnalyzer()
