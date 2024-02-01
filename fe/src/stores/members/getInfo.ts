import axios from 'axios';
import type { Member } from '$lib/types/member';
import { writable, type Writable } from 'svelte/store';

export type DataExportFormat = 'json' | 'csv';

export const memberInfo: Member = {
  memberName: '',
  webfinger: '',
  displayName: { String: '', Valid: false },
  email: '',
  profile_pic: '',
  bio: { String: '', Valid: false },
  matrix: { String: '', Valid: false },
  xmpp: { String: '', Valid: false },
  irc: { String: '', Valid: false },
  homepage: { String: '', Valid: false },
  regdate: 0,
  roles: [],
  visibility: "private",
  followers_uri: '',
  following_uri: '',
  sessionTimeout: { Int64: 0, Valid: false },
  active: false,
  uuid: ''
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
      const res = await axios.get(`/api/members/${email_or_username}/info`, {
        headers: {
          'Authorization': `Bearer ${jwtToken}`
        }
      });
      if (res.data.message !== "success") {
        throw new Error('Error while retrieving member data');
      }
      const member: Member = res.data.data;
      console.debug('member data retrieved from API: ', member);
      return member;
    },
    exportData: async (input: DataExportRequest) => {
      return new Promise<FileResponse>(async (resolve, reject) => {
        exportState.set('loading');

        const res = await axios.get(`/api/members/export/${input.target}`, {
          headers: {
            Authorization: `Bearer ${input.jwtToken}`
          },
          responseType: 'blob'
        });
        if (res.status === 200) {
          const fileName = res.headers['content-disposition']
            ?.split('filename=')[1]
            ?.replace(/['"]/g, '');

          const fileBlob: FileResponse = new Blob([res.data], {
            type: res.headers['content-type'],
          }) as FileResponse;

          exportState.set('success');
          resolve(fileBlob)
        } else {
          exportState.set('error');
          reject(res.data);
        }
      });
    },
  }
}

export const memberStore: MemberStore = createMemberStore();
