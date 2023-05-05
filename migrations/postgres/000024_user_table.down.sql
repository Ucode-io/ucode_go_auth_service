ALTER TABLE "user" DROP COLUMN if EXISTS language_id UUID REFERENCES "language"("id");;
ALTER TABLE "user" DROP COLUMN if EXISTS timezone_id UUID REFERENCES "timezone"("id");;
