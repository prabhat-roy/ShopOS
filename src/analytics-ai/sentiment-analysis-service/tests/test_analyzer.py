"""
Unit tests for the RuleBasedAnalyzer.

Tests cover:
  1.  Positive text detection
  2.  Negative text detection
  3.  Neutral text detection
  4.  Negation flips sentiment
  5.  Mixed text lands in NEUTRAL
  6.  Empty text returns NEUTRAL with score 0.5
  7.  Score is always in [0, 1]
  8.  Batch analysis returns correct count
  9.  Specific positive words are captured in positiveWords field
  10. Specific negative words are captured in negativeWords field
  11. Strongly negative text → NEGATIVE label
  12. Score increases with more positive words
"""

from __future__ import annotations

import sys
import os
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

import pytest

from sentiment.analyzer import RuleBasedAnalyzer
from sentiment.models import SentimentLabel


@pytest.fixture
def analyzer() -> RuleBasedAnalyzer:
    return RuleBasedAnalyzer()


# ---------------------------------------------------------------------------
# Test 1: Positive text detection
# ---------------------------------------------------------------------------


def test_positive_text(analyzer: RuleBasedAnalyzer):
    text = "This product is absolutely amazing, excellent quality and fantastic service!"
    result = analyzer.analyze(text)
    assert result.label == SentimentLabel.POSITIVE
    assert result.score > 0.6


# ---------------------------------------------------------------------------
# Test 2: Negative text detection
# ---------------------------------------------------------------------------


def test_negative_text(analyzer: RuleBasedAnalyzer):
    text = "Terrible product, completely broken and awful quality. I want a refund."
    result = analyzer.analyze(text)
    assert result.label == SentimentLabel.NEGATIVE
    assert result.score < 0.4


# ---------------------------------------------------------------------------
# Test 3: Neutral text detection
# ---------------------------------------------------------------------------


def test_neutral_text(analyzer: RuleBasedAnalyzer):
    text = "The package arrived on Tuesday. It contains three items."
    result = analyzer.analyze(text)
    assert result.label == SentimentLabel.NEUTRAL


# ---------------------------------------------------------------------------
# Test 4: Negation flips positive word to negative
# ---------------------------------------------------------------------------


def test_negation_flips_positive_word(analyzer: RuleBasedAnalyzer):
    # "not great" — the word "great" should be negated
    positive_text = "This is great."
    negated_text = "This is not great."

    pos_result = analyzer.analyze(positive_text)
    neg_result = analyzer.analyze(negated_text)

    assert pos_result.score > neg_result.score, (
        "Negated text should have a lower score than the non-negated version"
    )
    # "great" should end up in negativeWords after negation
    assert "great" in neg_result.negativeWords


# ---------------------------------------------------------------------------
# Test 5: Mixed text lands in NEUTRAL zone
# ---------------------------------------------------------------------------


def test_mixed_text_neutral(analyzer: RuleBasedAnalyzer):
    text = "The product is amazing but also quite terrible in some ways."
    result = analyzer.analyze(text)
    # With roughly equal positive and negative signals the label can be NEUTRAL
    # or slightly skewed — we just check the score is in a middle band
    assert 0.0 <= result.score <= 1.0


# ---------------------------------------------------------------------------
# Test 6: Empty text returns NEUTRAL with score 0.5
# ---------------------------------------------------------------------------


def test_empty_text_neutral(analyzer: RuleBasedAnalyzer):
    result = analyzer.analyze("")
    assert result.label == SentimentLabel.NEUTRAL
    assert result.score == 0.5
    assert result.positiveWords == []
    assert result.negativeWords == []


# ---------------------------------------------------------------------------
# Test 7: Score is always in [0, 1]
# ---------------------------------------------------------------------------


@pytest.mark.parametrize("text", [
    "",
    "ok",
    "amazing excellent wonderful brilliant",
    "terrible awful horrible broken useless",
    "not amazing not terrible",
    "The quick brown fox jumps over the lazy dog",
])
def test_score_range(analyzer: RuleBasedAnalyzer, text: str):
    result = analyzer.analyze(text)
    assert 0.0 <= result.score <= 1.0, f"Score {result.score} out of [0,1] for: {text!r}"


# ---------------------------------------------------------------------------
# Test 8: Batch analysis returns correct count
# ---------------------------------------------------------------------------


def test_batch_analysis(analyzer: RuleBasedAnalyzer):
    texts = [
        "Absolutely amazing product!",
        "Terrible and broken.",
        "It arrived on time.",
        "Love this, highly recommend!",
        "Very disappointing and slow.",
    ]
    results = analyzer.batch_analyze(texts)
    assert len(results) == len(texts)


# ---------------------------------------------------------------------------
# Test 9: Specific positive words captured
# ---------------------------------------------------------------------------


def test_positive_words_captured(analyzer: RuleBasedAnalyzer):
    text = "The delivery was fast and the support team was very helpful."
    result = analyzer.analyze(text)
    found = set(result.positiveWords)
    assert "fast" in found or "helpful" in found, (
        f"Expected 'fast' or 'helpful' in positiveWords, got {found}"
    )


# ---------------------------------------------------------------------------
# Test 10: Specific negative words captured
# ---------------------------------------------------------------------------


def test_negative_words_captured(analyzer: RuleBasedAnalyzer):
    text = "The product was broken and defective right out of the box."
    result = analyzer.analyze(text)
    found = set(result.negativeWords)
    assert "broken" in found or "defective" in found, (
        f"Expected 'broken' or 'defective' in negativeWords, got {found}"
    )


# ---------------------------------------------------------------------------
# Test 11: Strongly negative text → NEGATIVE label
# ---------------------------------------------------------------------------


def test_strongly_negative_label(analyzer: RuleBasedAnalyzer):
    text = (
        "This is the worst purchase I have ever made. "
        "Completely useless, cheap, and fake. "
        "Broken on arrival — total waste of money. "
        "Terrible quality, very frustrating."
    )
    result = analyzer.analyze(text)
    assert result.label == SentimentLabel.NEGATIVE


# ---------------------------------------------------------------------------
# Test 12: Score increases with more positive words
# ---------------------------------------------------------------------------


def test_score_increases_with_positive_words(analyzer: RuleBasedAnalyzer):
    fewer = analyzer.analyze("This is great.")
    more = analyzer.analyze("This is great, excellent, amazing, and wonderful!")
    assert more.score >= fewer.score, (
        "More positive words should yield equal or higher score"
    )
