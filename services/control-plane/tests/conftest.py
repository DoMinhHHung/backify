import os

os.environ["APP_ENV"] = "test"
os.environ["JWT_SECRET"] = "test-secret-that-is-long-enough-for-hs256-ok"
os.environ["JWT_EXPIRE_MINUTES"] = "60"
os.environ["INTERNAL_API_KEY"] = "test-internal-key-that-is-long-enough-ok"
os.environ["RABBITMQ_ENABLED"] = "false"
os.environ["DATABASE_URL"] = "postgresql://unused:unused@localhost:5432/unused"
