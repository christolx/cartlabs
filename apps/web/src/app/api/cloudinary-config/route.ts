import { NextResponse } from "next/server";

export function GET() {
  return NextResponse.json(
    {
      cloudName: process.env.NEXT_PUBLIC_CLOUDINARY_CLOUD_NAME ?? "",
      uploadPreset: process.env.NEXT_PUBLIC_CLOUDINARY_UPLOAD_PRESET ?? "",
    },
    { headers: { "Cache-Control": "private, max-age=300" } },
  );
}
