import type { Member } from '$lib/types/member';
import { writable, type Writable } from 'svelte/store';

export type DataExportFormat = 'json' | 'csv';

export const memberInfo: Member = {
  memberName: '',
  webfinger: '',
  displayName: '',
  email: '',
  profile_pic: '',
  bio: '',
  regdate: 0,
  roles: [],
  visibility: "private",
  followers_uri: '',
  following_uri: '',
  active: false,
  uuid: '',
  customFields: Array.from({ length: 0 }, () => new Map()),
};

export type DataExportRequest = {
  jwtToken: string;
  target: DataExportFormat;
};

interface FileResponse extends Blob {
  name: string;
};

export type ExportState = 'idle' | 'loading' | 'success' | 'error';

interface MemberStore extends Writable<Member> {
  getMember: (jwtToken: string, email_or_username: string) => Promise<Member>;
  exportData: (input: DataExportRequest) => Promise<FileResponse>;
}

function createMemberStore(): MemberStore {
  const exportState = writable<ExportState>('idle');

  const { subscribe, set, update } = writable<Member>(memberInfo);
  return {
    subscribe,
    set,
    update,
    getMember: async (jwtToken: string, email_or_username: string) => {
      const res = await fetch(`/api/members/${email_or_username}/info`, {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${jwtToken}`
        }
      });
      if (!res.ok) {
        throw new Error('Error while retrieving member data');
      }
      const resData = await res.json();

      if (resData.message !== "success") {
        throw new Error('Error while retrieving member data');
      }
      const member: Member = resData.data;
      console.debug('member data retrieved from API: ', member);
      return member;
    },
    exportData: async (input: DataExportRequest) => {
      return new Promise<FileResponse>(async (resolve, reject) => {
        exportState.set('loading');

        try {
          const res = await fetch(`/api/members/export/${input.target}`, {
            method: 'GET',
            headers: {
              Authorization: `Bearer ${input.jwtToken}`
            }
          });

          if (res.ok) {
            const fileName = res.headers.get('content-disposition')
              ?.split('filename=')[1]
              ?.replace(/['"]/g, '');

            const fileBlob: Blob = await res.blob();
            (fileBlob as FileResponse).name = fileName || 'export';

            const fileResponse = fileBlob as FileResponse;

            exportState.set('success');
            resolve(fileResponse);
          } else {
            const errorData = await res.json();
            exportState.set('error');
            reject(errorData);
          }
        } catch (error) {
          exportState.set('error');
          reject(error);
        }
      });
    },
  }
}

export const memberStore: MemberStore = createMemberStore();
