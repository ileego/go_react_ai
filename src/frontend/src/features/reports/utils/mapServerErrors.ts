import type { UseFormSetError, FieldValues, Path } from 'react-hook-form'

interface ServerError {
  field: string
  message: string
}

export function mapServerErrors<T extends FieldValues>(
  errors: ServerError[],
  setError: UseFormSetError<T>
) {
  for (const err of errors) {
    setError(err.field as Path<T>, {
      type: 'manual',
      message: err.message,
    })
  }
}
