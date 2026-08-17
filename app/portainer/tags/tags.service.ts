import axios, { parseAxiosError } from '../services/axios/axios';

import { Tag, TagId } from './types';

export async function getTags() {
  try {
    // Axios adds the shared /api base URL and authentication configuration.
    const { data } = await axios.get<Tag[]>(buildUrl());
    return data;
  } catch (err) {
    throw parseAxiosError(err as Error, 'Unable to retrieve tags');
  }
}

export async function createTag(name: string) {
  try {
    // The object becomes the JSON body expected by Go's tagCreatePayload.
    const { data: tag } = await axios.post<Tag>(buildUrl(), { name });
    return tag;
  } catch (err) {
    throw parseAxiosError(err as Error, 'Unable to create tag');
  }
}

export async function deleteTag(id: TagId) {
  try {
    await axios.delete(buildUrl(id));
  } catch (err) {
    throw parseAxiosError(err as Error, 'Unable to delte tag');
  }
}

function buildUrl(id?: TagId) {
  let url = '/tags';
  if (id) {
    url += `/${id}`;
  }

  return url;
}
