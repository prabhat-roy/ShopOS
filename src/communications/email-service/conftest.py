"""Root conftest.py — ensures the project root is on sys.path for pytest."""
import sys
import os

# Add the service root to sys.path so that `import app.*` resolves correctly
sys.path.insert(0, os.path.dirname(__file__))
