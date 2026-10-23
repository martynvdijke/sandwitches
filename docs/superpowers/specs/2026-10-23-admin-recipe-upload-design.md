# Admin recipe upload from the Android app

**Date:** 2026-10-23
**Status:** Approved

## Goal

Let staff and superusers create a recipe, including an optional photo, from the
Android app. Non-admins must not see the action, and the API must reject them
even if they call the endpoint directly.

## Decisions

- Upload scope is recipe details plus a photo: title, description, ingredients,
  instructions, servings, price, tags and an optional image.
- Enforcement is in both places: `/api/v1/me` exposes the staff flags, the app
  hides the entry point, and `POST /api/v1/recipes` requires an admin.
- The entry point is a floating "Add recipe" button on Home, visible only to
  admins. It is not a new tab and not a button in the Admin queue.
- Staff-created recipes are approved immediately (`is_approved = true`).
- The existing recipe endpoint accepts JSON and multipart on the same route.
  Multipart adds the optional `image` field; there is no second upload request.

## Backend

1. `internal/api/dto.go`: add `is_staff` and `is_superuser` to `UserDTO` and
   populate them in `userToDTO`.
2. `internal/api/api.go`:
   - `createRecipe` returns 403 unless `user.IsStaff || user.IsSuperuser`.
   - JSON requests keep the existing binding behaviour.
   - Multipart requests parse the recipe fields with `PostForm`, accept tags as
     repeated `tags` fields or a comma-separated `tags`/`tags_string` value,
     deduplicate tag names, and save an optional `image` with the existing
     `utils.SaveUploadedFile` helper under `media/recipes`.
   - The image is saved before the database row so a failed upload does not
     leave a half-created recipe. Bad numeric fields or an unreadable image
     return 400.
   - The created recipe gets `uploaded_by_id`, `is_approved = true`, and returns
     the normal recipe JSON with the image path.
3. `internal/api/openapi.go`: document both request content types, the binary
   `image` field and the new user flags.
4. Tests cover admin JSON creation, admin multipart creation with a small PNG,
   multipart without an image, non-admin rejection, invalid multipart input and
   the `/me` flags.

## Android app

1. `mobile/lib/core/roles.dart` provides the shared `isAdmin(me)` check against
   `is_staff` and `is_superuser`. The Admin queue uses it too.
2. `ApiRepository.createRecipe` sends one multipart request with a longer
   send/receive timeout and surfaces the server's error message.
3. `RecipeCreateScreen` is a full-screen form outside the bottom-navigation
   shell at `/admin/recipes/new`. It validates title, servings and price, lets
   the user pick or remove a photo, keeps the form on failure and shows API
   errors.
4. On success it invalidates the Home recipe list, recipe of the day and tags,
   then replaces the form with the new recipe detail.
5. A non-admin opening the route directly sees an "Admins only" message.
6. Widget tests cover the role helper, Home button visibility, validation,
   successful submission and API errors. The live API E2E test also creates a
   recipe with a photo through `ApiRepository.createRecipe`.

## Error handling

- 403 from the API is displayed in the form and is the authoritative check.
- 400 validation errors are displayed using the API's `message` field.
- Uploads use a 60-second timeout so photos on slow connections do not fail at
  the default 10 seconds.
- If the image cannot be written, the backend returns 400 before creating the
  recipe row.

## Verification

- `go test ./...`
- `go vet ./...`
- `flutter analyze`
- `flutter test`
- Optional live run: `mobile/tool/run_live_e2e.sh`
