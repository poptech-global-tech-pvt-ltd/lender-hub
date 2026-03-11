-- lending-hub-service/internal/infrastructure/postgres/migrations/0001_create_enums.sql

-- Onboarding status
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'lender_onboarding_status') THEN
        CREATE TYPE lender_onboarding_status AS ENUM (
          'PENDING',
          'IN_PROGRESS',
          'SUCCESS',
          'INELIGIBLE',
          'FAILED'
        );
    END IF;
END$$;

-- Profile status
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'lender_profile_status') THEN
        CREATE TYPE lender_profile_status AS ENUM (
          'NOT_STARTED',
          'IN_PROGRESS',
          'ACTIVE',
          'INELIGIBLE',
          'BLOCKED'
        );
    END IF;
END$$;

-- Payment status
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'lender_payment_status') THEN
        CREATE TYPE lender_payment_status AS ENUM (
          'PENDING',
          'SUCCESS',
          'FAILED',
          'REFUNDED',
          'EXPIRED',
          'CANCELLED'
        );
    END IF;
END$$;

-- Refund status
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'lender_refund_status') THEN
        CREATE TYPE lender_refund_status AS ENUM (
          'PENDING',
          'SUCCESS',
          'FAILED'
        );
    END IF;
END$$;

-- Idempotency status
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'idempotency_status') THEN
        CREATE TYPE idempotency_status AS ENUM (
          'PROCESSING',
          'COMPLETED',
          'FAILED'
        );
    END IF;
END$$;

-- Supporting enums

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'onboarding_step') THEN
        CREATE TYPE onboarding_step AS ENUM (
          'USER_DATA',
          'EMI_SELECTION',
          'KYC',
          'KFS',
          'MITC',
          'AUTOPAY'
        );
    END IF;
END$$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'onboarding_step_status') THEN
        CREATE TYPE onboarding_step_status AS ENUM (
          'PENDING',
          'SUCCESS',
          'FAILED'
        );
    END IF;
END$$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'channel_type') THEN
        CREATE TYPE channel_type AS ENUM (
          'WEB',
          'ANDROID',
          'IOS'
        );
    END IF;
END$$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'request_source') THEN
        CREATE TYPE request_source AS ENUM (
          'CHECKOUT',
          'PDP',
          'PLP',
          'CART',
          'CX'
        );
    END IF;
END$$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'refund_reason') THEN
        CREATE TYPE refund_reason AS ENUM (
          'USER_CANCELLED',
          'PRODUCT_RETURN',
          'ORDER_CANCELLED'
        );
    END IF;
END$$;
