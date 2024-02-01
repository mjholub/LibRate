import type { Member } from '$lib/types/member';

export type DataExportFormat = 'json' | 'csv' | 'sql';

export type DataExportRequest = {
  jwtToken: string;
  target: DataExportFormat;
}

interface MemberStore extends Writable<Member> {
  getMember: (jwtToken: string, email_or_username: string) => Promise<Member>;
  exportData: (input: DataExportRequest) => Promise<File>;
}

function createMemberStore(): MemberStore {
  const { subscribe, set, update } = writable<Member>({} as Member);
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
      return new Promise<File>(async (resolve, reject) => {
        const res = await axios.get(`/api/members/export/${input.target}`, {
          headers: {
            Authorization: `Bearer ${input.jwtToken}`
          },
          responseType: 'blob'
        });
        res.status === 200 ? resolve(res.data) : reject(res.data);
      });
    },
  }

  export const memberStore: MemberStore = createMemberStore();
